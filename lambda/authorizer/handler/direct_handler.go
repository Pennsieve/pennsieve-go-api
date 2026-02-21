package handler

import (
	"context"
	"fmt"

	coreAuthorizer "github.com/pennsieve/pennsieve-go-core/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
)

// DirectAuthorizeRequest is the request payload for direct Lambda-to-Lambda invocation.
// Callers provide a user ID and optionally an organization and/or dataset ID.
type DirectAuthorizeRequest struct {
	UserID         int64  `json:"user_id"`
	OrganizationID int64  `json:"organization_id,omitempty"`
	DatasetID      string `json:"dataset_id,omitempty"`
}

// DirectAuthorizeResponse is the response payload for direct Lambda-to-Lambda invocation.
type DirectAuthorizeResponse struct {
	IsAuthorized bool                   `json:"is_authorized"`
	Claims       map[string]interface{} `json:"claims,omitempty"`
	Error        string                 `json:"error,omitempty"`
}

// DirectHandler handles direct Lambda-to-Lambda invocation for authorization.
// It looks up the user by internal ID (no JWT required) and builds claims based
// on which IDs are provided:
//   - user_id only: returns user claim
//   - user_id + organization_id: returns user, organization, and team claims
//   - user_id + organization_id + dataset_id: returns user, organization, dataset, and team claims
func DirectHandler(ctx context.Context, request DirectAuthorizeRequest) (DirectAuthorizeResponse, error) {
	logger := log.WithFields(log.Fields{
		"user_id":         request.UserID,
		"organization_id": request.OrganizationID,
		"dataset_id":      request.DatasetID,
	})
	logger.Info("direct authorizer request")

	if request.UserID == 0 {
		return DirectAuthorizeResponse{
			IsAuthorized: false,
			Error:        "user_id is required",
		}, nil
	}

	if request.DatasetID != "" && request.OrganizationID == 0 {
		return DirectAuthorizeResponse{
			IsAuthorized: false,
			Error:        "organization_id is required when dataset_id is provided",
		}, nil
	}

	// Open Pennsieve DB Connection
	db, err := pgdb.ConnectRDS()
	if err != nil {
		logger.WithError(err).Error("unable to connect to RDS instance")
		return DirectAuthorizeResponse{IsAuthorized: false}, err
	}
	defer db.Close()
	postgresDB := pgdb.New(db)

	// Look up the user by internal ID
	currentUser, err := postgresDB.GetUserById(ctx, request.UserID)
	if err != nil {
		logger.WithError(err).Error("unable to get user by ID")
		return DirectAuthorizeResponse{
			IsAuthorized: false,
			Error:        fmt.Sprintf("unable to get user: %s", err.Error()),
		}, nil
	}

	claims := map[string]interface{}{
		coreAuthorizer.LabelUserClaim: &user.Claim{
			Id:           currentUser.Id,
			NodeId:       currentUser.NodeId,
			IsSuperAdmin: currentUser.IsSuperAdmin,
		},
	}

	// If organization_id is provided, add organization and team claims
	if request.OrganizationID != 0 {
		orgClaim, err := postgresDB.GetOrganizationClaim(ctx, currentUser.Id, request.OrganizationID)
		if err != nil {
			logger.WithError(err).Error("unable to get organization claim")
			return DirectAuthorizeResponse{
				IsAuthorized: false,
				Error:        fmt.Sprintf("unable to get organization claim: %s", err.Error()),
			}, nil
		}
		claims[coreAuthorizer.LabelOrganizationClaim] = orgClaim

		teamClaims, err := postgresDB.GetTeamClaims(ctx, currentUser.Id)
		if err != nil {
			logger.WithError(err).Error("unable to get team claims")
			return DirectAuthorizeResponse{
				IsAuthorized: false,
				Error:        fmt.Sprintf("unable to get team claims: %s", err.Error()),
			}, nil
		}
		claims[coreAuthorizer.LabelTeamClaims] = teamClaims
	}

	// If dataset_id is provided, add dataset claim (organization_id is guaranteed present)
	if request.DatasetID != "" {
		datasetClaim, err := postgresDB.GetDatasetClaim(ctx, currentUser, request.DatasetID, request.OrganizationID)
		if err != nil {
			logger.WithError(err).Error("unable to get dataset claim")
			return DirectAuthorizeResponse{
				IsAuthorized: false,
				Error:        fmt.Sprintf("unable to get dataset claim: %s", err.Error()),
			}, nil
		}
		if datasetClaim.Role == role.None {
			return DirectAuthorizeResponse{
				IsAuthorized: false,
				Error:        "user has no access to dataset",
			}, nil
		}
		claims[coreAuthorizer.LabelDatasetClaim] = datasetClaim
	}

	return DirectAuthorizeResponse{
		IsAuthorized: true,
		Claims:       claims,
	}, nil
}
