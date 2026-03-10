package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/pennsieve/pennsieve-go-api/authorizer/helpers"
	coreAuthorizer "github.com/pennsieve/pennsieve-go-core/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
)

// CallbackValidateRequest is the payload sent to the service's validator Lambda.
type CallbackValidateRequest struct {
	CallbackToken  string `json:"callbackToken"`
	ExecutionRunID string `json:"executionRunId"`
}

// CallbackValidateResponse is the expected response from the service's validator Lambda.
// The validator returns node IDs; the authorizer resolves them to full claims via Postgres.
type CallbackValidateResponse struct {
	IsAuthorized       bool   `json:"isAuthorized"`
	UserNodeID         string `json:"userNodeId"`
	OrganizationNodeID string `json:"organizationNodeId"`
	DatasetNodeID      string `json:"datasetNodeId"`
	Error              string `json:"error,omitempty"`
}

// getValidatorArn looks up the Lambda ARN for the given service name from environment variables.
// Environment variable format: CALLBACK_VALIDATOR_<SERVICE_NAME_UPPERCASED_WITH_UNDERSCORES>
// e.g., CALLBACK_VALIDATOR_WORKFLOW_SERVICE for "workflow-service"
func getValidatorArn(serviceName string) (string, error) {
	envKey := "CALLBACK_VALIDATOR_" + strings.ToUpper(strings.ReplaceAll(serviceName, "-", "_"))
	arn := os.Getenv(envKey)
	if arn == "" {
		return "", fmt.Errorf("no callback validator configured for service: %s", serviceName)
	}
	return arn, nil
}

// handleCallbackAuth handles requests with Callback authorization.
// It parses the header, invokes the appropriate service's validator Lambda to verify the token
// and get node IDs, then resolves those node IDs to full claims via Postgres.
func handleCallbackAuth(ctx context.Context, event events.APIGatewayV2CustomAuthorizerV2Request) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	logger := log.WithFields(log.Fields{"authType": "callback"})

	callbackAuth, err := helpers.ParseCallbackAuth(event.Headers["authorization"])
	if err != nil {
		logger.Error(err)
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: false,
		}, nil
	}

	logger = logger.WithFields(log.Fields{
		"service":        callbackAuth.Service,
		"executionRunId": callbackAuth.ExecutionRunID,
	})

	// Invoke the service's validator Lambda
	validateResp, err := invokeValidator(ctx, logger, callbackAuth)
	if err != nil {
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: false,
		}, nil
	}

	if !validateResp.IsAuthorized {
		logger.WithField("error", validateResp.Error).Warn("callback token validation failed")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: false,
		}, nil
	}

	// Resolve node IDs to full claims via Postgres (same pattern as DirectHandler)
	db, err := pgdb.ConnectRDS()
	if err != nil {
		logger.WithError(err).Error("unable to connect to RDS instance")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: false,
		}, err
	}
	defer db.Close()
	postgresDB := pgdb.New(db)

	currentUser, err := getUserByNodeId(ctx, db, validateResp.UserNodeID)
	if err != nil {
		logger.WithError(err).Error("unable to get user by node ID")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: false,
		}, nil
	}

	claims := map[string]interface{}{
		coreAuthorizer.LabelUserClaim: &user.Claim{
			Id:           currentUser.Id,
			NodeId:       currentUser.NodeId,
			IsSuperAdmin: currentUser.IsSuperAdmin,
		},
	}

	orgClaim, err := postgresDB.GetOrganizationClaimByNodeId(ctx, currentUser.Id, validateResp.OrganizationNodeID)
	if err != nil {
		logger.WithError(err).Error("unable to get organization claim")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: false,
		}, nil
	}
	claims[coreAuthorizer.LabelOrganizationClaim] = orgClaim

	datasetClaim, err := postgresDB.GetDatasetClaim(ctx, currentUser, validateResp.DatasetNodeID, orgClaim.IntId)
	if err != nil {
		logger.WithError(err).Error("unable to get dataset claim")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: false,
		}, nil
	}
	if datasetClaim.Role == role.None {
		logger.Warn("user has no access to dataset")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: false,
		}, nil
	}
	claims[coreAuthorizer.LabelDatasetClaim] = datasetClaim

	logger.Info("callback token authorization successful")
	return events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: true,
		Context:      claims,
	}, nil
}

func invokeValidator(ctx context.Context, logger *log.Entry, callbackAuth *helpers.CallbackAuth) (*CallbackValidateResponse, error) {
	validatorArn, err := getValidatorArn(callbackAuth.Service)
	if err != nil {
		logger.Error(err)
		return nil, err
	}

	payload, err := json.Marshal(CallbackValidateRequest{
		CallbackToken:  callbackAuth.Token,
		ExecutionRunID: callbackAuth.ExecutionRunID,
	})
	if err != nil {
		logger.WithError(err).Error("failed to marshal validate request")
		return nil, err
	}

	cfg, err := awsconfig.LoadDefaultConfig(ctx)
	if err != nil {
		logger.WithError(err).Error("unable to load AWS config")
		return nil, err
	}

	lambdaClient := lambda.NewFromConfig(cfg)
	result, err := lambdaClient.Invoke(ctx, &lambda.InvokeInput{
		FunctionName: aws.String(validatorArn),
		Payload:      payload,
	})
	if err != nil {
		logger.WithError(err).Error("failed to invoke validator Lambda")
		return nil, err
	}

	var validateResp CallbackValidateResponse
	if err := json.Unmarshal(result.Payload, &validateResp); err != nil {
		logger.WithError(err).Error("failed to unmarshal validator response")
		return nil, err
	}

	return &validateResp, nil
}