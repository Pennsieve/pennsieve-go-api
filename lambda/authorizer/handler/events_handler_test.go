package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test/mocks"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/teamUser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testNow = time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)

type fakeApps struct {
	hasAccess      bool
	err            error
	appUuid        string
	tokenWorkspace *manager.TokenWorkspace
	calls          int
}

func (f *fakeApps) CanAccessApp(_ context.Context, _ *pgdbModels.User, tokenWorkspace *manager.TokenWorkspace, appUuid string) (bool, error) {
	f.calls++
	f.appUuid, f.tokenWorkspace = appUuid, tokenWorkspace
	return f.hasAccess, f.err
}

// harness builds an eventsAuthorizer whose token validation returns the
// params' JWT (expiring in expiresIn) and whose session uses the params' mocks.
type harness struct {
	params   *mocks.ClaimsManagerParams
	user     *pgdbModels.User
	apps     *fakeApps
	sessions int
	closed   int
	auth     eventsAuthorizer
}

func newHarness(t *testing.T, params *mocks.ClaimsManagerParams, expiresIn time.Duration) *harness {
	h := &harness{params: params, user: test.NewUser(101, 1001), apps: &fakeApps{}}
	params.WithUserQueryMocked(t, h.user)
	require.NoError(t, params.TestJWT.Token.Set(jwt.ExpirationKey, testNow.Add(expiresIn)))
	h.auth = eventsAuthorizer{
		validateToken: func(b []byte) (jwt.Token, error) {
			if string(b) != "valid-token" {
				return nil, errors.New("bad token")
			}
			return params.TestJWT.Token, nil
		},
		openSession: func(context.Context, jwt.Token) (*eventsSession, error) {
			h.sessions++
			return &eventsSession{claims: params.BuildClaimsManager(), apps: h.apps, close: func() { h.closed++ }}, nil
		},
		now: func() time.Time { return testNow },
	}
	return h
}

func request(operation, namespace, channel string) EventsAuthorizerRequest {
	return EventsAuthorizerRequest{
		AuthorizationToken: "valid-token",
		RequestContext:     EventsRequestContext{RequestId: "r1", Operation: operation, ChannelNamespaceName: namespace, Channel: channel},
	}
}

func allowed(ttl int) EventsAuthorizerResponse {
	return EventsAuthorizerResponse{IsAuthorized: true, TTLOverride: ttl}
}

var denied = EventsAuthorizerResponse{IsAuthorized: false, TTLOverride: 0}

func TestEventsConnect(t *testing.T) {
	h := newHarness(t, mocks.NewClaimsManagerParams(t), time.Hour)
	assert.Equal(t, allowed(eventsMaxAuthTTL), h.auth.authorize(context.Background(), request(eventConnect, "", "")))
	assert.Zero(t, h.sessions, "connect needs no database")

	bearer := request(eventConnect, "", "")
	bearer.AuthorizationToken = "Bearer valid-token"
	assert.Equal(t, allowed(eventsMaxAuthTTL), h.auth.authorize(context.Background(), bearer))

	for _, token := range []string{"", "  ", "Bearer ", "other-token"} {
		r := request(eventConnect, "", "")
		r.AuthorizationToken = token
		assert.Equal(t, denied, h.auth.authorize(context.Background(), r), "token %q", token)
	}
}

func TestEventsTTLNeverOutlivesToken(t *testing.T) {
	h := newHarness(t, mocks.NewClaimsManagerParams(t), 42*time.Second)
	assert.Equal(t, allowed(42), h.auth.authorize(context.Background(), request(eventConnect, "", "")))
}

func TestEventsPublishAndUnknownOperationsDenied(t *testing.T) {
	h := newHarness(t, mocks.NewClaimsManagerParams(t), time.Hour)
	assert.Equal(t, denied, h.auth.authorize(context.Background(), request(eventPublish, "datasets", "/datasets/"+testUUID)))
	assert.Equal(t, denied, h.auth.authorize(context.Background(), request("EVENT_SOMETHING", "", "")))
	assert.Zero(t, h.sessions)
}

func TestEventsInvalidChannelDeniedWithoutDatabase(t *testing.T) {
	h := newHarness(t, mocks.NewClaimsManagerParams(t), time.Hour)
	assert.Equal(t, denied, h.auth.authorize(context.Background(), request(eventSubscribe, "datasets", "/datasets/*")))
	assert.Zero(t, h.sessions)
}

func TestEventsSessionFailureDenied(t *testing.T) {
	h := newHarness(t, mocks.NewClaimsManagerParams(t), time.Hour)
	h.auth.openSession = func(context.Context, jwt.Token) (*eventsSession, error) { return nil, errors.New("connection refused") }
	assert.Equal(t, denied, h.auth.authorize(context.Background(), request(eventSubscribe, "datasets", "/datasets/"+testUUID)))
}

func TestEventsSubscribeDataset(t *testing.T) {
	datasetNodeId := "N:dataset:" + testUUID
	const orgId = int64(7)

	for name, c := range map[string]struct {
		role role.Role
		want EventsAuthorizerResponse
	}{
		"viewer":  {role.Viewer, allowed(eventsMaxAuthTTL)},
		"no role": {role.None, denied},
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t, mocks.NewClaimsManagerParams(t), time.Hour)
			pg := h.params.MockPennsievePg
			pg.OnGetOrganizationIdForDataset(datasetNodeId).Return(orgId, nil)
			pg.OnGetOrganizationClaim(h.user.Id, orgId).Return(&organization.Claim{Role: pgdbModels.Read, IntId: orgId}, nil)
			pg.OnGetDatasetClaim(h.user, datasetNodeId, orgId).Return(&dataset.Claim{Role: c.role, NodeId: datasetNodeId}, nil)

			assert.Equal(t, c.want, h.auth.authorize(context.Background(), request(eventSubscribe, "datasets", "/datasets/"+testUUID)))
			assert.Equal(t, 1, h.closed, "session closed")
			h.params.AssertMockExpectations(t)
		})
	}
}

func TestEventsSubscribeDatasetLookupFailureDenied(t *testing.T) {
	h := newHarness(t, mocks.NewClaimsManagerParams(t), time.Hour)
	h.params.MockPennsievePg.OnGetOrganizationIdForDataset("N:dataset:"+testUUID).Return(int64(0), errors.New("timeout"))
	assert.Equal(t, denied, h.auth.authorize(context.Background(), request(eventSubscribe, "datasets", "/datasets/"+testUUID)))
}

func TestEventsSubscribeWorkspaceRuns(t *testing.T) {
	orgNodeId := "N:organization:" + testUUID
	for name, c := range map[string]struct {
		permission pgdbModels.DbPermission
		want       EventsAuthorizerResponse
	}{
		"member":     {pgdbModels.Read, allowed(eventsMaxAuthTTL)},
		"not member": {pgdbModels.NoPermission, denied},
	} {
		for _, channel := range []string{"/runs/org-" + testUUID + "/*", "/runs/org-" + testUUID + "/" + testUUID2} {
			t.Run(name+" "+channel, func(t *testing.T) {
				h := newHarness(t, mocks.NewClaimsManagerParams(t), time.Hour)
				pg := h.params.MockPennsievePg
				pg.OnGetOrganizationClaimByNodeId(h.user.Id, orgNodeId).Return(&organization.Claim{Role: c.permission, IntId: 7, NodeId: orgNodeId}, nil)
				pg.OnGetTeamClaimsForOrg(h.user.Id, int64(7)).Return([]teamUser.Claim{}, nil).Maybe()

				assert.Equal(t, c.want, h.auth.authorize(context.Background(), request(eventSubscribe, "runs", channel)))
			})
		}
	}
}

func TestEventsSubscribeWorkspaceRunsAPITokenForAnotherWorkspace(t *testing.T) {
	params := mocks.NewClaimsManagerParams(t).WithTokenWorkspace(t, manager.TokenWorkspace{Id: 3, NodeId: "N:organization:" + testUUID2})
	h := newHarness(t, params, time.Hour)
	assert.Equal(t, denied, h.auth.authorize(context.Background(), request(eventSubscribe, "runs", "/runs/org-"+testUUID+"/*")))
}

func TestEventsSubscribeUserRuns(t *testing.T) {
	h := newHarness(t, mocks.NewClaimsManagerParams(t), time.Hour)
	own := h.user.NodeId[len("N:user:"):]

	assert.Equal(t, allowed(eventsMaxAuthTTL), h.auth.authorize(context.Background(), request(eventSubscribe, "runs", "/runs/user-"+own+"/*")))
	assert.Equal(t, allowed(eventsMaxAuthTTL), h.auth.authorize(context.Background(), request(eventSubscribe, "runs", "/runs/user-"+own+"/"+testUUID2)))
	assert.Equal(t, denied, h.auth.authorize(context.Background(), request(eventSubscribe, "runs", "/runs/user-"+testUUID+"/*")),
		"another user's runs")
}

func TestEventsSubscribeApplication(t *testing.T) {
	channel := "/applications/" + testUUID

	t.Run("allowed", func(t *testing.T) {
		h := newHarness(t, mocks.NewClaimsManagerParams(t), time.Hour)
		h.apps.hasAccess = true
		assert.Equal(t, allowed(eventsMaxAuthTTL), h.auth.authorize(context.Background(), request(eventSubscribe, "applications", channel)))
		assert.Equal(t, testUUID, h.apps.appUuid)
		assert.Nil(t, h.apps.tokenWorkspace)
	})
	t.Run("denied", func(t *testing.T) {
		h := newHarness(t, mocks.NewClaimsManagerParams(t), time.Hour)
		assert.Equal(t, denied, h.auth.authorize(context.Background(), request(eventSubscribe, "applications", channel)))
	})
	t.Run("check failed", func(t *testing.T) {
		h := newHarness(t, mocks.NewClaimsManagerParams(t), time.Hour)
		h.apps.hasAccess, h.apps.err = true, errors.New("throttled")
		assert.Equal(t, denied, h.auth.authorize(context.Background(), request(eventSubscribe, "applications", channel)))
	})
	t.Run("API token passes its workspace", func(t *testing.T) {
		tw := manager.TokenWorkspace{Id: 3, NodeId: "N:organization:" + testUUID2}
		h := newHarness(t, mocks.NewClaimsManagerParams(t).WithTokenWorkspace(t, tw), time.Hour)
		h.apps.hasAccess = true
		assert.Equal(t, allowed(eventsMaxAuthTTL), h.auth.authorize(context.Background(), request(eventSubscribe, "applications", channel)))
		assert.Equal(t, &tw, h.apps.tokenWorkspace)
	})
}

func TestEventsRequestShape(t *testing.T) {
	// Example event from the AppSync Events docs.
	raw := `{
	  "authorizationToken": "ExampleAUTHtoken123123123",
	  "requestContext": {
	    "apiId": "aaaaaa123123123example123",
	    "accountId": "111122223333",
	    "requestId": "f4081827-1111-4444-5555-5cf4695f339f",
	    "operation": "EVENT_SUBSCRIBE",
	    "channelNamespaceName": "news",
	    "channel": "/news/latest"
	  },
	  "requestHeaders": {"header": "value"}
	}`
	var r EventsAuthorizerRequest
	require.NoError(t, json.Unmarshal([]byte(raw), &r))
	assert.Equal(t, "ExampleAUTHtoken123123123", r.AuthorizationToken)
	assert.Equal(t, EventsRequestContext{
		ApiId: "aaaaaa123123123example123", AccountId: "111122223333", RequestId: "f4081827-1111-4444-5555-5cf4695f339f",
		Operation: eventSubscribe, ChannelNamespaceName: "news", Channel: "/news/latest",
	}, r.RequestContext)

	out, err := json.Marshal(eventsDeny())
	require.NoError(t, err)
	assert.JSONEq(t, `{"isAuthorized": false, "ttlOverride": 0}`, string(out))
}
