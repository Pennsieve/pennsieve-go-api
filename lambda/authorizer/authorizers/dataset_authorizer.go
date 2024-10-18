package authorizers

import (
	"context"
	"errors"
	"fmt"

	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset/role"
	log "github.com/sirupsen/logrus"
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
		log.Error("unable to get current user")
		return nil, err
	}

	// Get Active Org
	orgInt := claimsManager.GetActiveOrg(ctx, currentUser)

	// Get Workspace Claim
	orgClaim, err := claimsManager.GetOrgClaim(ctx, currentUser, orgInt)
	if err != nil {
		log.Error("unable to get Organization Role")
		return nil, err
	}

	// Get Dataset Claim
	datasetClaim, err := claimsManager.GetDatasetClaim(ctx, currentUser, d.DatasetId, orgInt)
	if err != nil {
		log.Error("unable to get Dataset Role")
		return nil, err
	}
	// If user has no role on provided dataset --> return
	if datasetClaim.Role == role.None {
		log.Error("user has no access to dataset")
		return nil, errors.New("user has no access to dataset")
	}

	// Get User Claim
	userClaim := claimsManager.GetUserClaim(ctx, currentUser)

	if authorizerMode == "LEGACY" {
		// Get Publisher's Claim
		teamClaims, err := claimsManager.GetTeamClaims(ctx, currentUser)
		if err != nil {
			log.Error(fmt.Sprintf("unable to get Team Claims for user: %d organization: %d",
				currentUser.Id, orgInt))
			return nil, err
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
