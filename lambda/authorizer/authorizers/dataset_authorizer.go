package authorizers

import (
	"context"
	"errors"
	"fmt"

	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset/role"
)

type DatasetAuthorizer struct {
	DatasetId string
}

func NewDatasetAuthorizer(datasetId string) Authorizer {
	return &DatasetAuthorizer{datasetId}
}

func (d *DatasetAuthorizer) GenerateClaims(ctx context.Context, claimsManager manager.IdentityManager, authorizerMode string) (map[string]interface{}, error) {
	// Get current user
	currentUser, err := claimsManager.GetCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get current user: %w", err)
	}

	// Get Active Org
	orgInt := claimsManager.GetActiveOrg(ctx, currentUser)

	// Get Workspace Claim
	orgClaim, err := claimsManager.GetOrgClaim(ctx, currentUser, orgInt)
	if err != nil {
		return nil, fmt.Errorf("unable to get Organization Role: %w", err)
	}

	// Get Dataset Claim
	datasetClaim, err := claimsManager.GetDatasetClaim(ctx, currentUser, d.DatasetId, orgInt)
	if err != nil {
		return nil, fmt.Errorf("unable to get Dataset Role: %w", err)
	}
	// If user has no role on provided dataset --> return
	if datasetClaim.Role == role.None {
		return nil, errors.New("user has no access to dataset")
	}

	// Get User Claim
	userClaim := claimsManager.GetUserClaim(ctx, currentUser)

	if authorizerMode == "LEGACY" {
		// Get Publisher's Claim
		teamClaims, err := claimsManager.GetTeamClaims(ctx, currentUser)
		if err != nil {
			return nil, fmt.Errorf("unable to get Team Claims for user: %d organization: %d: %w",
				currentUser.Id, orgInt, err)
		}

		return map[string]interface{}{
			"user_claim":    userClaim,
			"org_claim":     orgClaim,
			"dataset_claim": datasetClaim,
			"teams_claim":   teamClaims,
		}, nil
	}

	return map[string]interface{}{
		"user_claim":    userClaim,
		"org_claim":     orgClaim,
		"dataset_claim": datasetClaim,
	}, nil
}
