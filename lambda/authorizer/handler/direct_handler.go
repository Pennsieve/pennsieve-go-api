package handler

import (
	"context"
	"database/sql"
	"fmt"

	coreAuthorizer "github.com/pennsieve/pennsieve-go-core/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
)

// DirectAuthorizeRequest is the request payload for direct Lambda-to-Lambda invocation.
// Callers provide node IDs for the user and optionally an organization and/or dataset.
type DirectAuthorizeRequest struct {
	UserNodeID         string `json:"user_node_id"`
	OrganizationNodeID string `json:"organization_node_id,omitempty"`
	DatasetNodeID      string `json:"dataset_node_id,omitempty"`
}

// getUserByNodeId looks up a Pennsieve user by their node ID (e.g. "N:user:...").
func getUserByNodeId(ctx context.Context, db *sql.DB, nodeId string) (*pgdbModels.User, error) {
	queryStr := "SELECT id, node_id, email, first_name, last_name, is_super_admin, COALESCE(preferred_org_id, -1) as preferred_org_id " +
		"FROM pennsieve.users WHERE node_id=$1;"

	var u pgdbModels.User
	row := db.QueryRowContext(ctx, queryStr, nodeId)
	err := row.Scan(
		&u.Id,
		&u.NodeId,
		&u.Email,
		&u.FirstName,
		&u.LastName,
		&u.IsSuperAdmin,
		&u.PreferredOrg)

	if err != nil {
		return nil, err
	}
	return &u, nil
}

// DirectAuthorizeResponse is the response payload for direct Lambda-to-Lambda invocation.
type DirectAuthorizeResponse struct {
	IsAuthorized bool                   `json:"is_authorized"`
	Claims       map[string]interface{} `json:"claims,omitempty"`
	Error        string                 `json:"error,omitempty"`
}

// DirectHandler handles direct Lambda-to-Lambda invocation for authorization.
// It looks up the user by node ID (no JWT required) and builds claims based
// on which node IDs are provided:
//   - user_node_id only: returns user claim
//   - user_node_id + organization_node_id: returns user, organization, and team claims
//   - user_node_id + organization_node_id + dataset_node_id: returns user, organization, dataset, and team claims
func DirectHandler(ctx context.Context, request DirectAuthorizeRequest) (DirectAuthorizeResponse, error) {
	logger := log.WithFields(log.Fields{
		"user_node_id":         request.UserNodeID,
		"organization_node_id": request.OrganizationNodeID,
		"dataset_node_id":      request.DatasetNodeID,
	})
	logger.Info("direct authorizer request")

	if request.UserNodeID == "" {
		return DirectAuthorizeResponse{
			IsAuthorized: false,
			Error:        "user_node_id is required",
		}, nil
	}

	if request.DatasetNodeID != "" && request.OrganizationNodeID == "" {
		return DirectAuthorizeResponse{
			IsAuthorized: false,
			Error:        "organization_node_id is required when dataset_node_id is provided",
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

	// Look up the user by node ID
	currentUser, err := getUserByNodeId(ctx, db, request.UserNodeID)
	if err != nil {
		logger.WithError(err).Error("unable to get user by node ID")
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

	// If organization_node_id is provided, add organization and team claims
	if request.OrganizationNodeID != "" {
		orgClaim, err := postgresDB.GetOrganizationClaimByNodeId(ctx, currentUser.Id, request.OrganizationNodeID)
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

	// If dataset_node_id is provided, add dataset claim (organization_node_id is guaranteed present)
	if request.DatasetNodeID != "" {
		// Use the org's integer ID from the org claim for the dataset schema lookup
		orgClaim := claims[coreAuthorizer.LabelOrganizationClaim]
		orgIntId := orgClaim.(*organization.Claim).IntId

		datasetClaim, err := postgresDB.GetDatasetClaim(ctx, currentUser, request.DatasetNodeID, orgIntId)
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
