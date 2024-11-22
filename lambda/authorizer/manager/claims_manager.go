package manager

import (
	"context"
	"errors"
	"os"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/teamUser"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/dydb"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	pgdbQueries "github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
)

type IdentityManager interface {
	GetCurrentUser(context.Context) (*pgdbModels.User, error)
	GetActiveOrg(context.Context, *pgdbModels.User) int64
	GetUserClaim(context.Context, *pgdbModels.User) user.Claim
	GetDatasetClaim(context.Context, *pgdbModels.User, string, int64) (*dataset.Claim, error)
	GetOrgClaim(context.Context, *pgdbModels.User, int64) (*organization.Claim, error)
	GetTeamClaims(context.Context, *pgdbModels.User) ([]teamUser.Claim, error)
	GetDatasetID(context.Context, string) (*string, error)
	GetUserTokenWorkspace() (UserTokenWorkspace, bool)
}

type ClaimsManager struct {
	PostgresDB    *pgdbQueries.Queries
	DynamoDB      *dydb.Queries
	Token         jwt.Token
	TokenClientID string
}

func NewClaimsManager(postgresDB *pgdbQueries.Queries, dynamoDB *dydb.Queries, token jwt.Token, tokenClientID string) IdentityManager {
	return &ClaimsManager{postgresDB, dynamoDB, token, tokenClientID}
}

func (c *ClaimsManager) GetDatasetClaim(ctx context.Context, currentUser *pgdbModels.User, datasetId string, orgInt int64) (*dataset.Claim, error) {
	datasetClaim, err := c.PostgresDB.GetDatasetClaim(ctx, currentUser, datasetId, orgInt)
	if err != nil {
		return nil, err
	}

	return datasetClaim, nil
}

func (c *ClaimsManager) GetDatasetID(ctx context.Context, manifestID string) (*string, error) {
	table := os.Getenv("MANIFEST_TABLE")

	manifest, err := c.DynamoDB.GetManifestById(ctx, table, manifestID)
	if err != nil {
		// log.Error("manifest could not be found")
		return nil, err
	}

	return &manifest.DatasetNodeId, nil
}

func (c *ClaimsManager) GetUserClaim(ctx context.Context, currentUser *pgdbModels.User) user.Claim {
	return user.Claim{
		Id:           currentUser.Id,
		NodeId:       currentUser.NodeId,
		IsSuperAdmin: currentUser.IsSuperAdmin,
	}
}

func (c *ClaimsManager) GetOrgClaim(ctx context.Context, currentUser *pgdbModels.User, orgInt int64) (*organization.Claim, error) {
	orgClaim, err := c.PostgresDB.GetOrganizationClaim(ctx, currentUser.Id, orgInt)
	if err != nil {
		return nil, err
	}

	return orgClaim, nil
}

func (c *ClaimsManager) GetTeamClaims(ctx context.Context, currentUser *pgdbModels.User) ([]teamUser.Claim, error) {
	teamClaims, err := c.PostgresDB.GetTeamClaims(ctx, currentUser.Id)
	if err != nil {
		return nil, err
	}

	return teamClaims, nil
}

func (c *ClaimsManager) GetUserTokenWorkspace() (UserTokenWorkspace, bool) {
	var workspace UserTokenWorkspace
	if jwtOrgId, hasKey := c.Token.Get("custom:organization_id"); !hasKey {
		return workspace, false
	} else {
		workspace.Id = jwtOrgId.(int64)
	}
	if jwtOrgNodeId, hasKey := c.Token.Get("custom:organization_node_id"); !hasKey {
		return workspace, false
	} else {
		workspace.NodeId = jwtOrgNodeId.(string)
	}

	return workspace, true

}

func (c *ClaimsManager) GetActiveOrg(ctx context.Context, currentUser *pgdbModels.User) int64 {
	orgInt := currentUser.PreferredOrg
	jwtOrg, hasKey := c.Token.Get("custom:organization_id")
	if hasKey {
		orgInt = jwtOrg.(int64)
	}
	return orgInt
}

func (c *ClaimsManager) GetCurrentUser(ctx context.Context) (*pgdbModels.User, error) {
	// Get Cognito User ID
	cognitoUserName, hasKey := c.Token.Get("username")
	if !hasKey {
		return nil, errors.New("Unauthorized")
	}

	// Get Pennsieve User from User Table, or Token Table
	clientIdClaim, _ := c.Token.Get("client_id") // Key is present or method would have returned before.
	isFromTokenPool := clientIdClaim == c.TokenClientID
	return getUser(ctx, c.PostgresDB, cognitoUserName.(string), isFromTokenPool)
}

// getUser returns a Pennsieve user from a cognito ID.
func getUser(ctx context.Context, q *pgdb.Queries, cognitoId string, isFromTokenPool bool) (*pgdbModels.User, error) {

	if isFromTokenPool {
		//var token pgdbModels.Token
		currentUser, err := q.GetUserByCognitoId(ctx, cognitoId)
		if err != nil {
			log.Fatalln("unable to get user:", err)
		}
		return currentUser, nil

	} else {
		//var user pgdbModels.User
		currentUser, err := q.GetByCognitoId(ctx, cognitoId)
		if err != nil {
			log.Fatalln("unable to get user:", err)
		}
		return currentUser, nil
	}
}

type UserTokenWorkspace struct {
	Id     int64
	NodeId string
}
