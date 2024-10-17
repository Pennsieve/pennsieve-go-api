package authorizers

import (
	"context"
	"errors"

	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset/role"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	log "github.com/sirupsen/logrus"
)

type DatasetAuthorizer struct {
	DatasetId string
}

func NewDatasetAuthorizer(datasetId string) Authorizer {
	return &DatasetAuthorizer{datasetId}
}

func (d *DatasetAuthorizer) GenerateClaims(ctx context.Context, claimsManager manager.IdentityManager) (map[string]interface{}, error) {
	// Get current user
	currentUser, err := claimsManager.GetCurrentUser(ctx)
	if err != nil {
		log.Error("unable to get current user")
		return nil, err
	}
	// Get Active Org
	orgInt := currentUser.PreferredOrg
	jwtOrg, hasKey := claimsManager.GetToken().Get("custom:organization_id")
	if hasKey {
		orgInt = jwtOrg.(int64)
	}

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

	userClaim := user.Claim{
		Id:           currentUser.Id,
		NodeId:       currentUser.NodeId,
		IsSuperAdmin: currentUser.IsSuperAdmin,
	}

	// Get ORG Claim
	orgClaim, err := claimsManager.GetQueryHandle().GetOrganizationClaim(ctx, currentUser.Id, orgInt)
	if err != nil {
		log.Error("unable to get Organization Role")
		return nil, err
	}

	return map[string]interface{}{
		"user_claim":    userClaim,
		"org_claim":     orgClaim,
		"dataset_claim": datasetClaim,
	}, nil
}
