package authorizers

import (
	"context"
	"errors"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset/role"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/dydb"
	pgdbQueries "github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
)

// will be deprecated
type ManifestAuthorizer struct {
	CurrentUser    *pgdbModels.User
	Queries        *pgdbQueries.Queries
	IdentitySource []string
	Token          jwt.Token
}

func NewManifestAuthorizer(currentUser *pgdbModels.User, pddb *pgdbQueries.Queries, IdentitySource []string, token jwt.Token) Authorizer {
	return &ManifestAuthorizer{currentUser, pddb, IdentitySource, token}
}

func (m *ManifestAuthorizer) GenerateClaims(ctx context.Context) (map[string]interface{}, error) {
	// Get Active Org
	orgInt := m.CurrentUser.PreferredOrg
	jwtOrg, hasKey := m.Token.Get("custom:organization_id")
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

	manifest, err := qDyDb.GetManifestById(ctx, table, m.IdentitySource[1])
	if err != nil {
		log.Error("manifest could not be found")
		return nil, err
	}

	datasetNodeId := manifest.DatasetNodeId
	// Get DATASET Claim
	var datasetClaim *dataset.Claim
	datasetClaim, err = m.Queries.GetDatasetClaim(ctx, m.CurrentUser, datasetNodeId, orgInt)
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
		Id:           m.CurrentUser.Id,
		NodeId:       m.CurrentUser.NodeId,
		IsSuperAdmin: m.CurrentUser.IsSuperAdmin,
	}

	// Get ORG Claim
	orgClaim, err := m.Queries.GetOrganizationClaim(ctx, m.CurrentUser.Id, orgInt)
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
