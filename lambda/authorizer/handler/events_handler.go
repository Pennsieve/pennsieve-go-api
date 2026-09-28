package handler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
)

// AppSync Events operations passed to a Lambda authorizer.
const (
	eventConnect   = "EVENT_CONNECT"
	eventSubscribe = "EVENT_SUBSCRIBE"
	eventPublish   = "EVENT_PUBLISH"
)

// eventsMaxAuthTTL caps how long AppSync caches an allow, in seconds. Losing
// access therefore takes up to this long to stop new subscriptions.
const eventsMaxAuthTTL = 300

// EventsAuthorizerRequest is the event AppSync Events sends a Lambda
// authorizer. channel and channelNamespaceName are empty for EVENT_CONNECT.
type EventsAuthorizerRequest struct {
	AuthorizationToken string               `json:"authorizationToken"`
	RequestContext     EventsRequestContext `json:"requestContext"`
	RequestHeaders     map[string]string    `json:"requestHeaders"`
}

type EventsRequestContext struct {
	ApiId                string `json:"apiId"`
	AccountId            string `json:"accountId"`
	RequestId            string `json:"requestId"`
	Operation            string `json:"operation"`
	ChannelNamespaceName string `json:"channelNamespaceName"`
	Channel              string `json:"channel"`
}

// EventsAuthorizerResponse is the authorizer's answer. AppSync caches it for
// ttlOverride seconds per API, operation, channel and token; 0 means not
// cached.
type EventsAuthorizerResponse struct {
	IsAuthorized bool `json:"isAuthorized"`
	TTLOverride  int  `json:"ttlOverride"`
}

// EventsHandler is the Lambda authorizer for the AppSync Event API.
//
//   - EVENT_CONNECT: allowed for a valid Pennsieve access token from either
//     Cognito pool (web app or agent/API key).
//   - EVENT_SUBSCRIBE: allowed per channel:
//     /datasets/<uuid> for a caller with a role on the dataset;
//     /runs/org-<uuid>/... for a member of the workspace;
//     /runs/user-<uuid>/... for that user only;
//     /applications/<uuid> when app-deploy-service's check-app-access allows it.
//   - EVENT_PUBLISH: always denied. Publishing is IAM-only; the backends sign
//     their requests, so this Lambda is never asked unless the API is
//     misconfigured.
//
// Denials are never cached, so a user who is granted access can subscribe
// straight away and a failed database lookup is retried. Allows are cached for
// up to eventsMaxAuthTTL, and never past the token's expiry. The token and
// request headers are never logged.
func EventsHandler(ctx context.Context, request EventsAuthorizerRequest) (EventsAuthorizerResponse, error) {
	return defaultEventsAuthorizer.authorize(ctx, request), nil
}

var defaultEventsAuthorizer = eventsAuthorizer{
	validateToken: validateCognitoJWT,
	openSession:   openEventsSession,
	now:           time.Now,
}

type eventsAuthorizer struct {
	validateToken func([]byte) (jwt.Token, error)
	openSession   func(ctx context.Context, token jwt.Token) (*eventsSession, error)
	now           func() time.Time
}

// eventsSession holds what a subscribe decision needs; close releases it.
type eventsSession struct {
	claims manager.IdentityManager
	apps   appAccessChecker
	close  func()
}

func (a eventsAuthorizer) authorize(ctx context.Context, request EventsAuthorizerRequest) EventsAuthorizerResponse {
	rc := request.RequestContext
	logger := log.WithFields(log.Fields{
		"requestId": rc.RequestId,
		"operation": rc.Operation,
		"channel":   rc.Channel,
	})

	token := bearerToken(request.AuthorizationToken)
	if token == "" {
		logger.Warn("denied: missing token")
		return eventsDeny()
	}
	jwtToken, err := a.validateToken([]byte(token))
	if err != nil {
		logger.WithError(err).Warn("denied: invalid token")
		return eventsDeny()
	}

	switch rc.Operation {
	case eventConnect:
		return a.allow(jwtToken)

	case eventSubscribe:
		channel, err := parseEventsChannel(rc.ChannelNamespaceName, rc.Channel)
		if err != nil {
			logger.WithError(err).Warn("denied: invalid channel")
			return eventsDeny()
		}
		session, err := a.openSession(ctx, jwtToken)
		if err != nil {
			logger.WithError(err).Error("denied: unable to open session")
			return eventsDeny()
		}
		defer session.close()
		if err := authorizeSubscribe(ctx, session, channel); err != nil {
			entry := logger.WithError(err)
			if isIndeterminate(err) {
				entry.Error("denied: access could not be decided")
			} else {
				entry.Info("denied")
			}
			return eventsDeny()
		}
		return a.allow(jwtToken)

	case eventPublish:
		logger.Error("denied: publish must use IAM; check the namespace's publish auth mode")
		return eventsDeny()
	}
	logger.Warn("denied: unknown operation")
	return eventsDeny()
}

// authorizeSubscribe returns nil when the caller may subscribe to channel.
// Errors wrapped in authorizers.IndeterminateError mean the decision couldn't
// be made (a lookup failed); others are denials.
func authorizeSubscribe(ctx context.Context, session *eventsSession, channel eventsChannel) error {
	switch channel.namespace {
	case namespaceDatasets:
		_, err := authorizers.NewDatasetAuthorizer("N:dataset:"+channel.id).GenerateClaims(ctx, session.claims, "")
		return err

	case namespaceRuns:
		switch channel.scopeKind {
		case runScopeOrg:
			_, err := authorizers.NewWorkspaceAuthorizer("N:organization:"+channel.scopeId).GenerateClaims(ctx, session.claims, "")
			return err
		case runScopeUser:
			user, err := session.claims.GetCurrentUser(ctx)
			if err != nil {
				return fmt.Errorf("unable to get current user: %w", err)
			}
			if user.NodeId != "N:user:"+channel.scopeId {
				return errors.New("run scope belongs to another user")
			}
			return nil
		}

	case namespaceApplications:
		user, err := session.claims.GetCurrentUser(ctx)
		if err != nil {
			return fmt.Errorf("unable to get current user: %w", err)
		}
		var tokenWorkspace *manager.TokenWorkspace
		if tw, ok := session.claims.GetTokenWorkspace(); ok {
			tokenWorkspace = &tw
		}
		hasAccess, err := session.apps.CanAccessApp(ctx, user, tokenWorkspace, channel.id)
		if err != nil {
			return authorizers.NewIndeterminateError(fmt.Errorf("checking app access: %w", err))
		}
		if !hasAccess {
			return errors.New("user has no access to application")
		}
		return nil
	}
	return fmt.Errorf("no rule for channel namespace %q", channel.namespace)
}

func (a eventsAuthorizer) allow(token jwt.Token) EventsAuthorizerResponse {
	ttl := eventsMaxAuthTTL
	if left := int(token.Expiration().Sub(a.now()).Seconds()); left < ttl {
		ttl = max(left, 0)
	}
	return EventsAuthorizerResponse{IsAuthorized: true, TTLOverride: ttl}
}

func eventsDeny() EventsAuthorizerResponse {
	return EventsAuthorizerResponse{IsAuthorized: false, TTLOverride: 0}
}

// bearerToken accepts the token with or without a "Bearer " prefix.
func bearerToken(authorization string) string {
	token := strings.TrimSpace(authorization)
	if len(token) > len("bearer ") && strings.EqualFold(token[:len("bearer ")], "bearer ") {
		token = strings.TrimSpace(token[len("bearer "):])
	}
	return token
}

// openEventsSession connects to Postgres for one subscribe decision. The AWS
// config and Lambda client are only built if an app channel needs them.
func openEventsSession(ctx context.Context, token jwt.Token) (*eventsSession, error) {
	db, err := pgdb.ConnectRDS()
	if err != nil {
		return nil, authorizers.NewIndeterminateError(fmt.Errorf("connecting to postgres: %w", err))
	}
	apps := &lambdaAppAccess{
		functionName: os.Getenv("CHECK_APP_ACCESS_LAMBDA_NAME"),
		scope: func(ctx context.Context, userId, organizationId int64) (manager.AccessScope, error) {
			return manager.GetAccessScope(ctx, db, userId, organizationId)
		},
		invoker: func(ctx context.Context) (lambdaInvoker, error) {
			cfg, err := config.LoadDefaultConfig(ctx)
			if err != nil {
				return nil, fmt.Errorf("loading AWS config: %w", err)
			}
			return lambda.NewFromConfig(cfg), nil
		},
	}
	return &eventsSession{
		// Subscribe decisions never read manifests, so no DynamoDB client.
		claims: manager.NewClaimsManager(pgdb.New(db), nil, token, tokenClientID, manifestTableName),
		apps:   apps,
		close:  func() { _ = db.Close() },
	}, nil
}
