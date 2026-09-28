package handler

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeInvoker struct {
	input *lambda.InvokeInput
	out   *lambda.InvokeOutput
	err   error
}

func (f *fakeInvoker) Invoke(_ context.Context, in *lambda.InvokeInput, _ ...func(*lambda.Options)) (*lambda.InvokeOutput, error) {
	f.input = in
	return f.out, f.err
}

func newLambdaAppAccess(inv *fakeInvoker, scopeOrg *int64) *lambdaAppAccess {
	return &lambdaAppAccess{
		functionName: "prod-app-deploy-service-check-app-access-lambda-use1",
		scope: func(_ context.Context, userId, organizationId int64) (manager.AccessScope, error) {
			*scopeOrg = organizationId
			return manager.AccessScope{
				WorkspaceNodeIds: []string{"N:organization:a", "N:organization:b"},
				TeamNodeIds:      []string{"N:team:t"},
			}, nil
		},
		invoker: func(context.Context) (lambdaInvoker, error) { return inv, nil },
	}
}

var appUser = &pgdbModels.User{Id: 101, NodeId: "N:user:" + testUUID}

func TestCheckAppAccessRequest(t *testing.T) {
	inv := &fakeInvoker{out: &lambda.InvokeOutput{Payload: []byte(`{"hasAccess":true}`)}}
	var scopeOrg int64 = -1
	ok, err := newLambdaAppAccess(inv, &scopeOrg).CanAccessApp(context.Background(), appUser, nil, testUUID2)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Equal(t, int64(0), scopeOrg, "all workspaces without an API token")
	assert.Equal(t, "prod-app-deploy-service-check-app-access-lambda-use1", aws.ToString(inv.input.FunctionName))
	assert.JSONEq(t, `{
		"appUuid": "`+testUUID2+`",
		"userNodeId": "N:user:`+testUUID+`",
		"workspaceNodeIds": ["N:organization:a", "N:organization:b"],
		"teamNodeIds": ["N:team:t"]
	}`, string(inv.input.Payload))
}

func TestCheckAppAccessAPITokenScope(t *testing.T) {
	inv := &fakeInvoker{out: &lambda.InvokeOutput{Payload: []byte(`{"hasAccess":false}`)}}
	var scopeOrg int64
	ok, err := newLambdaAppAccess(inv, &scopeOrg).CanAccessApp(context.Background(), appUser, &manager.TokenWorkspace{Id: 3}, testUUID2)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, int64(3), scopeOrg)
}

func TestCheckAppAccessFailures(t *testing.T) {
	var scopeOrg int64
	for name, inv := range map[string]*fakeInvoker{
		"invoke error":   {err: errors.New("throttled")},
		"function error": {out: &lambda.InvokeOutput{FunctionError: aws.String("Unhandled"), Payload: []byte(`{}`)}},
		"bad payload":    {out: &lambda.InvokeOutput{Payload: []byte(`not json`)}},
	} {
		t.Run(name, func(t *testing.T) {
			ok, err := newLambdaAppAccess(inv, &scopeOrg).CanAccessApp(context.Background(), appUser, nil, testUUID2)
			assert.Error(t, err)
			assert.False(t, ok)
		})
	}

	t.Run("function not configured", func(t *testing.T) {
		l := newLambdaAppAccess(&fakeInvoker{}, &scopeOrg)
		l.functionName = ""
		ok, err := l.CanAccessApp(context.Background(), appUser, nil, testUUID2)
		assert.Error(t, err)
		assert.False(t, ok)
	})
}

func TestCheckAppAccessResponseIgnoresExtraFields(t *testing.T) {
	var resp checkAppAccessResponse
	require.NoError(t, json.Unmarshal([]byte(`{"hasAccess":true,"accessType":"workspace"}`), &resp))
	assert.True(t, resp.HasAccess)
}
