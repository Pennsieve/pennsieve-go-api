package authorizers

import (
	"context"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset/role"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/dydb"
	log "github.com/sirupsen/logrus"
)

// will be deprecated
type ManifestAuthorizer struct {
	ManifestID string
}

func NewManifestAuthorizer(manifestId string) Authorizer {
	return &ManifestAuthorizer{manifestId}
}

func (m *ManifestAuthorizer) GenerateClaims(ctx context.Context, claimsManager manager.IdentityManager) (map[string]interface{}, error) {
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

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		log.Error("unable to load SDK config")
		return nil, err
	}

	// Create an Amazon DynamoDB client.
	client := dynamodb.NewFromConfig(cfg)
	table := os.Getenv("MANIFEST_TABLE")
	qDyDb := dydb.New(client)

	manifest, err := qDyDb.GetManifestById(ctx, table, m.ManifestID)
	if err != nil {
		log.Error("manifest could not be found")
		return nil, err
	}

	datasetNodeId := manifest.DatasetNodeId
	// Get DATASET Claim
	datasetClaim, err := claimsManager.GetDatasetClaim(ctx, currentUser, datasetNodeId, orgInt)
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
