package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	coreAuthorizer "github.com/pennsieve/pennsieve-go-core/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/dydb"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
)

// WebSocketHandler is the entry point for the WebSocket Lambda REQUEST authorizer.
//
// Why a separate handler from `Handler`:
//
// The existing `Handler` accepts `APIGatewayV2CustomAuthorizerV2Request` — the
// HTTP API V2 payload format 2.0. API Gateway *WebSocket* REQUEST authorizers
// can only deliver payload format 1.0 (REST-style event shape) — there is no
// AWS-side toggle to make WebSocket emit V2 payload (confirmed via AWS docs and
// aws-cdk #14858). The two payload shapes differ enough (no `IdentitySource`
// array, different `requestContext`, different return-type expectations) that a
// purely-shape-aware dispatcher inside one handler would be more error-prone
// than splitting.
//
// What this authorizer does:
//
//  1. Pulls the access token from `?token=` in the connection URL. Browser
//     WebSocket clients cannot set custom headers, so the token MUST be in the
//     query string — this is why "token in URL" is the canonical WS auth
//     pattern, even though it leaks the token into CloudWatch access logs.
//     Mitigations are standard: short-lived tokens (~1h), wss-only.
//
//  2. Validates the JWT using the SAME `validateCognitoJWT` the HTTP authorizer
//     uses — single source of truth for "what is a valid Pennsieve JWT."
//
//  3. Resolves the Cognito sub to a Pennsieve user node ID via the SAME
//     `ClaimsManager.GetCurrentUser` the HTTP authorizer uses. This is the
//     non-trivial bit that consumers cannot reproduce without postgres access:
//     the JWT's `sub`/`username` is the Cognito ID, and the Pennsieve user
//     `node_id` is a separate UUID linked via `users.cognito_id`.
//
//  4. If `datasetId` is present in the query string, runs the same dataset role
//     check the HTTP `DatasetAuthorizer` runs. Refuses with role.None.
//
//  5. If `computeNodeId` is present in the query string (and we have a resolved
//     org node ID from the claims), invokes account-service's check-access
//     Lambda. Refuses if hasAccess=false. This is the one cross-service call
//     the authorizer makes — see the package doc + Terraform IAM grant for the
//     rationale. The check is opt-in via identity source; services that don't
//     care about compute-node access simply omit `computeNodeId` from the
//     handshake URL.
//
//  6. Returns a V1-shaped IAM policy + flattened context. Context fields are
//     scalars only (V1 limitation — no nested objects); we serialize the
//     user/org/dataset claims as JSON strings so consumers can unmarshal if
//     they want the full shape, plus break out `userNodeId` / `orgNodeId` /
//     `datasetRole` / `computeNodeAccess` as direct scalar fields for
//     convenience.
//
// What this authorizer does NOT do:
//
//   - Per-route authorization. The IAM policy resource ARN uses a wildcard so
//     all routes on this WebSocket connection get one auth decision; that
//     matches how WebSocket REQUEST authorizers work — they only fire at
//     `$connect`, never on subsequent message frames.
//
// Required query-string parameters:
//
//	token         — Cognito access token (always required)
//	datasetId     — Pennsieve dataset node ID, format N:dataset:<uuid> (optional;
//	                presence triggers dataset role check)
//	orgId         — Pennsieve organization node ID, format N:organization:<uuid>
//	                (optional; required if computeNodeId is set, since
//	                check-access is org-scoped)
//	computeNodeId — Plain UUID of a Pennsieve compute node (optional; presence
//	                triggers cross-service call to account-service check-access)
//
// Missing/invalid parameters produce a Deny policy with a `errorReason` context
// field so callers can log the cause without exposing it to the WebSocket
// client (the client only sees a 401 / 403 from API Gateway).
func WebSocketHandler(ctx context.Context, event events.APIGatewayCustomAuthorizerRequestTypeRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	logger := log.WithFields(log.Fields{
		"methodArn":             event.MethodArn,
		"queryStringParameters": redactToken(event.QueryStringParameters),
	})
	logger.Info("WebSocket REQUEST authorizer invoked")

	token := event.QueryStringParameters["token"]
	if token == "" {
		logger.Warn("rejecting — missing token query parameter")
		return denyResponse(event.MethodArn, "missing_token"), nil
	}

	jwtToken, err := validateCognitoJWT([]byte(token))
	if err != nil {
		logger.WithError(err).Warn("rejecting — JWT invalid")
		return denyResponse(event.MethodArn, "invalid_token"), nil
	}

	db, err := pgdb.ConnectRDS()
	if err != nil {
		logger.WithError(err).Error("postgres connect failed")
		// Non-nil error → API Gateway returns 500 to the client. JWT was valid
		// but we can't enforce platform-level access without the DB.
		return events.APIGatewayCustomAuthorizerResponse{}, err
	}
	defer db.Close()
	postgresDB := pgdb.New(db)

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.WithError(err).Error("aws config load failed")
		return events.APIGatewayCustomAuthorizerResponse{}, err
	}
	dynamoDB := dydb.New(dynamodb.NewFromConfig(cfg))

	claimsManager := manager.NewClaimsManager(postgresDB, dynamoDB, jwtToken, tokenClientID, manifestTableName)
	authorizerMode := os.Getenv("AUTHORIZER_MODE")

	// Pick the authorizer based on what identity sources the client supplied.
	// chat-service always sends datasetId → DatasetAuthorizer (most restrictive,
	// also resolves user + org claims as a side effect). Other future WS
	// services might send only orgId or just the token.
	var auth authorizers.Authorizer
	switch {
	case event.QueryStringParameters["datasetId"] != "":
		auth = authorizers.NewDatasetAuthorizer(event.QueryStringParameters["datasetId"])
	case event.QueryStringParameters["orgId"] != "":
		auth = authorizers.NewWorkspaceAuthorizer(event.QueryStringParameters["orgId"])
	default:
		auth = authorizers.NewUserAuthorizer()
	}

	claims, err := auth.GenerateClaims(ctx, claimsManager, authorizerMode)
	if err != nil {
		// Includes "user has no access to dataset" from DatasetAuthorizer.
		logger.WithError(err).Warn("rejecting — claims generation failed")
		return denyResponse(event.MethodArn, fmt.Sprintf("claims_failed: %s", err.Error())), nil
	}

	// Optional cross-service compute-node access check. Only runs when the
	// caller actually has a compute node concept (chat-service does;
	// hypothetical future notification/telemetry services may not). The
	// `accessType` ("owner" / "shared" / "workspace" / "team") gets flattened
	// into the response context for consumers who want to display it.
	var computeNodeAccessType string
	if nodeUUID := event.QueryStringParameters["computeNodeId"]; nodeUUID != "" {
		userNodeID := extractPrincipalID(claims)
		orgNodeID := extractOrgNodeID(claims)
		if orgNodeID == "" {
			logger.Warn("rejecting — computeNodeId provided without an org claim")
			return denyResponse(event.MethodArn, "compute_node_check_missing_org"), nil
		}
		accessType, err := checkComputeNodeAccess(ctx, cfg, userNodeID, nodeUUID, orgNodeID)
		if err != nil {
			logger.WithError(err).Error("compute-node access check failed")
			// Fail closed on transport errors — if we can't confirm access, refuse.
			return denyResponse(event.MethodArn, fmt.Sprintf("compute_node_check_failed: %s", err.Error())), nil
		}
		if accessType == "" {
			logger.WithFields(log.Fields{"user": userNodeID, "node": nodeUUID, "org": orgNodeID}).
				Warn("rejecting — compute-node access denied")
			return denyResponse(event.MethodArn, "compute_node_access_denied"), nil
		}
		computeNodeAccessType = accessType
	}

	return allowResponseWithComputeNode(event.MethodArn, claims, computeNodeAccessType), nil
}

// allowResponseWithComputeNode builds an Allow IAM policy authorizing the
// caller for ALL routes on this WebSocket API (wildcard route key), with an
// optional `computeNodeAccess` field added to the response context when the
// caller ran a compute-node access check.
//
// WebSocket REQUEST authorizers fire only on $connect, and the resulting
// connection inherits the policy for its lifetime — there is no re-evaluation
// on subsequent message frames.
func allowResponseWithComputeNode(methodArn string, claims map[string]interface{}, computeNodeAccessType string) events.APIGatewayCustomAuthorizerResponse {
	ctx := flattenContext(claims)
	if computeNodeAccessType != "" {
		ctx["computeNodeAccess"] = computeNodeAccessType
	}
	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: extractPrincipalID(claims),
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{{
				Action:   []string{"execute-api:Invoke"},
				Effect:   "Allow",
				Resource: []string{wildcardArn(methodArn)},
			}},
		},
		Context: ctx,
	}
}

// denyResponse returns an explicit Deny policy. API Gateway translates this to
// 403 Forbidden on the WebSocket handshake. We never return `Unauthorized`
// directly — the explicit Deny path lets us thread a redacted error reason
// through the context map for log correlation.
func denyResponse(methodArn, reason string) events.APIGatewayCustomAuthorizerResponse {
	return events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: "unauthorized",
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
			Statement: []events.IAMPolicyStatement{{
				Action:   []string{"execute-api:Invoke"},
				Effect:   "Deny",
				Resource: []string{methodArn},
			}},
		},
		Context: map[string]interface{}{"errorReason": reason},
	}
}

// flattenContext turns the nested `claims` map produced by the per-type
// authorizers into V1-shaped flat scalar context fields. WebSocket REQUEST
// authorizers (payload v1.0) DO NOT support nested context values — only
// strings, numbers, and booleans. So we:
//
//  1. Pull the most common fields out as direct scalars (userNodeId,
//     orgNodeId, datasetRole) for ergonomic access in consumers.
//  2. JSON-serialize each top-level claim object as a string under
//     userClaim / orgClaim / datasetClaim / teamClaims — consumers that
//     need the full shape can `json.Unmarshal` them.
//
// Consumers (e.g. chat-service $connect handler) thus get:
//
//	authMap["userNodeId"] = "N:user:..."     // string scalar
//	authMap["orgNodeId"]  = "N:organization:..."
//	authMap["datasetRole"] = "viewer"        // role.String()
//	authMap["userClaim"]  = "{...}"          // JSON string of full claim
func flattenContext(claims map[string]interface{}) map[string]interface{} {
	out := map[string]interface{}{}

	if v, ok := claims[coreAuthorizer.LabelUserClaim]; ok {
		if uc, ok := v.(*user.Claim); ok && uc != nil {
			out["userNodeId"] = uc.NodeId
			if b, err := json.Marshal(uc); err == nil {
				out["userClaim"] = string(b)
			}
		}
	}
	if v, ok := claims[coreAuthorizer.LabelOrganizationClaim]; ok {
		if oc, ok := v.(*organization.Claim); ok && oc != nil {
			out["orgNodeId"] = oc.NodeId
			if b, err := json.Marshal(oc); err == nil {
				out["orgClaim"] = string(b)
			}
		}
	}
	if v, ok := claims[coreAuthorizer.LabelDatasetClaim]; ok {
		if dc, ok := v.(*dataset.Claim); ok && dc != nil {
			out["datasetNodeId"] = dc.NodeId
			out["datasetRole"] = dc.Role.String()
			if b, err := json.Marshal(dc); err == nil {
				out["datasetClaim"] = string(b)
			}
		}
	}
	if v, ok := claims[coreAuthorizer.LabelTeamClaims]; ok {
		if b, err := json.Marshal(v); err == nil {
			out["teamClaims"] = string(b)
		}
	}
	return out
}

// extractPrincipalID returns the user node ID for use as the IAM policy
// PrincipalId field. Falls back to "unknown" if claims are malformed (the
// authorize call already succeeded by this point, so this is purely defensive).
func extractPrincipalID(claims map[string]interface{}) string {
	if v, ok := claims[coreAuthorizer.LabelUserClaim]; ok {
		if uc, ok := v.(*user.Claim); ok && uc != nil {
			return uc.NodeId
		}
	}
	return "unknown"
}

// extractOrgNodeID returns the organization node ID from the resolved claims,
// or empty string if no org claim was generated (UserAuthorizer path doesn't
// produce one in non-LEGACY mode). The compute-node access check requires this
// because account-service's check-access Lambda is org-scoped.
func extractOrgNodeID(claims map[string]interface{}) string {
	if v, ok := claims[coreAuthorizer.LabelOrganizationClaim]; ok {
		if oc, ok := v.(*organization.Claim); ok && oc != nil {
			return oc.NodeId
		}
	}
	return ""
}

// wildcardArn converts a route-specific methodArn (ending in `/$connect`)
// into one that grants access to all routes on the API/stage:
//
//	arn:aws:execute-api:us-east-1:123:abc/dev/$connect
//	→
//	arn:aws:execute-api:us-east-1:123:abc/dev/*
//
// We use this in the Allow policy so the single $connect authorization
// implicitly authorizes the resulting WebSocket session for every message
// route. (WebSocket REQUEST authorizers only fire at $connect; there's no
// way to re-authorize mid-session, so per-route policies wouldn't be enforced
// anyway.)
func wildcardArn(methodArn string) string {
	// Replace the trailing route segment with `*`. methodArn format is:
	//   arn:aws:execute-api:{region}:{account}:{apiId}/{stage}/{route}
	// The last `/` separates stage and route.
	for i := len(methodArn) - 1; i >= 0; i-- {
		if methodArn[i] == '/' {
			return methodArn[:i+1] + "*"
		}
	}
	return methodArn
}

// redactToken returns a shallow copy of the query string map with the `token`
// value replaced. The full token must not appear in CloudWatch logs even at
// debug level — it's a bearer credential while valid.
func redactToken(params map[string]string) map[string]string {
	if params == nil {
		return nil
	}
	out := make(map[string]string, len(params))
	for k, v := range params {
		if k == "token" {
			if v == "" {
				out[k] = ""
			} else {
				out[k] = "[redacted]"
			}
			continue
		}
		out[k] = v
	}
	return out
}

