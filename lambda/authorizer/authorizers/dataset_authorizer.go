package authorizers

import (
	"context"
	"errors"

	"github.com/lestrrat-go/jwx/v2/jwt"

	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset/role"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	pgdbQueries "github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
)

type DatasetAuthorizer struct {
	CurrentUser    *pgdbModels.User
	Queries        *pgdbQueries.Queries
	IdentitySource []string
	Token          jwt.Token
}

func NewDatasetAuthorizer(currentUser *pgdbModels.User, pddb *pgdbQueries.Queries, IdentitySource []string, token jwt.Token) Authorizer {
	return &DatasetAuthorizer{currentUser, pddb, IdentitySource, token}
}

func (d *DatasetAuthorizer) GenerateClaims(ctx context.Context) (map[string]interface{}, error) {
	// Get Active Org
	orgInt := d.CurrentUser.PreferredOrg
	jwtOrg, hasKey := d.Token.Get("custom:organization_id")
	if hasKey {
		orgInt = jwtOrg.(int64)
	}

	// Get DATASET Claim
	var datasetClaim *dataset.Claim
	datasetClaim, err := d.Queries.GetDatasetClaim(ctx, d.CurrentUser, d.IdentitySource[1], orgInt)
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
		Id:           d.CurrentUser.Id,
		NodeId:       d.CurrentUser.NodeId,
		IsSuperAdmin: d.CurrentUser.IsSuperAdmin,
	}

	// Get ORG Claim
	orgClaim, err := d.Queries.GetOrganizationClaim(ctx, d.CurrentUser.Id, orgInt)
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
