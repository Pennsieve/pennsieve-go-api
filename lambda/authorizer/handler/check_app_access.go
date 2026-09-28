package handler

// App access check used by the events authorizer for /applications/<appUuid>.
//
// App access rules live in app-deploy-service (CanAccessApp: public, owner, or
// a grant to the user, one of their workspaces or one of their teams), so the
// authorizer asks app-deploy-service's check-app-access Lambda instead of
// duplicating them, the same way the WebSocket authorizer asks account-service
// about compute nodes. Apps can be shared across workspaces, so the request
// carries all of the user's workspaces and teams.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
)

// checkAppAccessRequest is the check-app-access Lambda's input.
type checkAppAccessRequest struct {
	AppUuid          string   `json:"appUuid"`
	UserNodeId       string   `json:"userNodeId"`
	WorkspaceNodeIds []string `json:"workspaceNodeIds"`
	TeamNodeIds      []string `json:"teamNodeIds"`
}

// checkAppAccessResponse is its output. An unknown app is hasAccess=false.
type checkAppAccessResponse struct {
	HasAccess bool `json:"hasAccess"`
}

type lambdaInvoker interface {
	Invoke(ctx context.Context, params *lambda.InvokeInput, optFns ...func(*lambda.Options)) (*lambda.InvokeOutput, error)
}

// appAccessChecker decides whether a user may subscribe to an app's channel.
type appAccessChecker interface {
	CanAccessApp(ctx context.Context, user *pgdbModels.User, tokenWorkspace *manager.TokenWorkspace, appUuid string) (bool, error)
}

type lambdaAppAccess struct {
	functionName string
	scope        func(ctx context.Context, userId, organizationId int64) (manager.AccessScope, error)
	invoker      func(ctx context.Context) (lambdaInvoker, error)
}

// CanAccessApp returns false with a nil error for a clean denial, and an error
// when access couldn't be decided. Both deny the subscription.
func (l *lambdaAppAccess) CanAccessApp(ctx context.Context, user *pgdbModels.User, tokenWorkspace *manager.TokenWorkspace, appUuid string) (bool, error) {
	if l.functionName == "" {
		return false, errors.New("CHECK_APP_ACCESS_LAMBDA_NAME not set")
	}
	var organizationId int64
	if tokenWorkspace != nil {
		organizationId = tokenWorkspace.Id
	}
	scope, err := l.scope(ctx, user.Id, organizationId)
	if err != nil {
		return false, err
	}
	payload, err := json.Marshal(checkAppAccessRequest{
		AppUuid:          appUuid,
		UserNodeId:       user.NodeId,
		WorkspaceNodeIds: scope.WorkspaceNodeIds,
		TeamNodeIds:      scope.TeamNodeIds,
	})
	if err != nil {
		return false, fmt.Errorf("marshaling check-app-access request: %w", err)
	}

	invoker, err := l.invoker(ctx)
	if err != nil {
		return false, err
	}
	out, err := invoker.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(l.functionName),
		InvocationType: lambdatypes.InvocationTypeRequestResponse,
		Payload:        payload,
	})
	if err != nil {
		return false, fmt.Errorf("invoking check-app-access: %w", err)
	}
	if out.FunctionError != nil {
		return false, fmt.Errorf("check-app-access function error: %s", aws.ToString(out.FunctionError))
	}
	var resp checkAppAccessResponse
	if err := json.Unmarshal(out.Payload, &resp); err != nil {
		return false, fmt.Errorf("unmarshaling check-app-access response: %w", err)
	}
	return resp.HasAccess, nil
}
