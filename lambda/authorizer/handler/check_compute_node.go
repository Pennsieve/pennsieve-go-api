package handler

// Compute-node access check used by the WebSocket REQUEST authorizer.
//
// This is the one cross-service runtime dependency this authorizer Lambda
// has: it invokes account-service's check-access Lambda when the WebSocket
// handshake URL carries `?computeNodeId=...`. The IAM grant for this is in
// terraform/iam.tf (the authorizer Lambda role's policy resource list
// includes the check-access Lambda ARN pulled from account-service's remote
// state).
//
// Why account-service is invoked from here, not from the consuming service:
// see the long-form note at the top of websocket_handler.go. tl;dr — moving
// the check inline keeps the auth boundary at API Gateway (consistent with
// every other Pennsieve service), gets API Gateway's per-identity-source
// caching for free, and means consuming services (chat, future
// notifications, etc.) don't all have to plumb their own check-access
// invocation. The cost is a runtime call from pennsieve-go-api to
// account-service — operationally similar to the postgres call this
// authorizer already makes, but worth being explicit about as the first
// time pennsieve-go-api reaches "outward" into the platform.

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	lambdatypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
)

// checkUserNodeAccessRequest mirrors account-service's
// internal/handler/checkaccess.CheckUserNodeAccessRequest. JSON tags match
// the upstream Lambda's expectation (note: camelCase, which is inconsistent
// with our snake_case direct-authorizer — but we follow the upstream).
type checkUserNodeAccessRequest struct {
	UserNodeId     string `json:"userNodeId"`
	NodeUuid       string `json:"nodeUuid"`
	OrganizationId string `json:"organizationId"`
}

type checkUserNodeAccessResponse struct {
	HasAccess      bool   `json:"hasAccess"`
	AccessType     string `json:"accessType,omitempty"` // owner | shared | workspace | team
	AccessSource   string `json:"accessSource,omitempty"`
	TeamId         string `json:"teamId,omitempty"`
	NodeUuid       string `json:"nodeUuid"`
	UserNodeId     string `json:"userNodeId"`
	OrganizationId string `json:"organizationId"`
}

// checkComputeNodeAccess invokes account-service's check-access Lambda to
// verify that `userNodeID` has access to compute node `nodeUUID` within
// `orgNodeID`. Returns the access type ("owner", "shared", "workspace",
// "team") on success, empty string on a clean denial, and an error on any
// transport / configuration / unmarshaling failure.
//
// Fails closed: any non-nil error from this function MUST cause the caller
// to refuse the connection. We never want a misconfigured
// CHECK_ACCESS_LAMBDA_NAME or a transient Lambda failure to silently let
// unauthorized users through.
func checkComputeNodeAccess(ctx context.Context, cfg aws.Config, userNodeID, nodeUUID, orgNodeID string) (string, error) {
	name := os.Getenv("CHECK_ACCESS_LAMBDA_NAME")
	if name == "" {
		return "", errors.New("CHECK_ACCESS_LAMBDA_NAME env var not set")
	}
	if userNodeID == "" || nodeUUID == "" || orgNodeID == "" {
		return "", fmt.Errorf("missing identifier (user=%q node=%q org=%q)", userNodeID, nodeUUID, orgNodeID)
	}

	payload, err := json.Marshal(checkUserNodeAccessRequest{
		UserNodeId:     userNodeID,
		NodeUuid:       nodeUUID,
		OrganizationId: orgNodeID,
	})
	if err != nil {
		return "", fmt.Errorf("marshaling check-access request: %w", err)
	}

	client := lambda.NewFromConfig(cfg)
	out, err := client.Invoke(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(name),
		InvocationType: lambdatypes.InvocationTypeRequestResponse,
		Payload:        payload,
	})
	if err != nil {
		return "", fmt.Errorf("invoking check-access: %w", err)
	}
	if out.FunctionError != nil {
		return "", fmt.Errorf("check-access function error: %s — payload=%s", aws.ToString(out.FunctionError), string(out.Payload))
	}

	var resp checkUserNodeAccessResponse
	if err := json.Unmarshal(out.Payload, &resp); err != nil {
		return "", fmt.Errorf("unmarshaling check-access response: %w (raw=%s)", err, string(out.Payload))
	}
	if !resp.HasAccess {
		return "", nil // clean denial — caller distinguishes from transport error
	}

	// AccessType should always be present when HasAccess=true per the
	// upstream API contract, but defend against the empty case so callers
	// don't see "access granted but no access type".
	if resp.AccessType == "" {
		return "unknown", nil
	}
	return resp.AccessType, nil
}
