package manager

import (
	"context"
	"errors"
	"fmt"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/teamUser"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
)

type IdentityManager interface {
	GetCurrentUser(ctx context.Context) (*pgdbModels.User, error)
	GetActiveOrg(ctx context.Context, user *pgdbModels.User) int64
	GetUserClaim(ctx context.Context, user *pgdbModels.User) user.Claim
	GetDatasetClaim(ctx context.Context, user *pgdbModels.User, datasetId string, orgId int64) (*dataset.Claim, error)
	GetOrgClaim(ctx context.Context, userId int64, orgId int64) (*organization.Claim, error)
	GetOrgClaimByNodeId(ctx context.Context, userId int64, orgNodeId string) (*organization.Claim, error)
	GetTeamClaims(ctx context.Context, userId int64) ([]teamUser.Claim, error)
	GetDatasetID(ctx context.Context, manifestId string) (string, error)
	GetTokenWorkspace() (TokenWorkspace, bool)
}

type ClaimsManager struct {
	PostgresDB        PennsievePgAPI
	DynamoDB          PennsieveDyAPI
	Token             jwt.Token
	TokenClientID     string
	ManifestTableName string
}

func NewClaimsManager(postgresDB PennsievePgAPI, dynamoDB PennsieveDyAPI, token jwt.Token, tokenClientID string, manifestTable string) IdentityManager {
	return &ClaimsManager{postgresDB, dynamoDB, token, tokenClientID, manifestTable}
}

func (c *ClaimsManager) GetDatasetClaim(ctx context.Context, currentUser *pgdbModels.User, datasetId string, orgInt int64) (*dataset.Claim, error) {
	datasetClaim, err := c.PostgresDB.GetDatasetClaim(ctx, currentUser, datasetId, orgInt)
	if err != nil {
		return nil, err
	}

	return datasetClaim, nil
}

func (c *ClaimsManager) GetDatasetID(ctx context.Context, manifestID string) (string, error) {
	manifest, err := c.DynamoDB.GetManifestById(ctx, c.ManifestTableName, manifestID)
	if err != nil {
		return "", err
	}

	return manifest.DatasetNodeId, nil
}

func (c *ClaimsManager) GetUserClaim(ctx context.Context, currentUser *pgdbModels.User) user.Claim {
	return user.Claim{
		Id:           currentUser.Id,
		NodeId:       currentUser.NodeId,
		IsSuperAdmin: currentUser.IsSuperAdmin,
	}
}

func (c *ClaimsManager) GetOrgClaim(ctx context.Context, userId int64, orgId int64) (*organization.Claim, error) {
	orgClaim, err := c.PostgresDB.GetOrganizationClaim(ctx, userId, orgId)
	if err != nil {
		return nil, err
	}

	return orgClaim, nil
}

func (c *ClaimsManager) GetOrgClaimByNodeId(ctx context.Context, userId int64, orgNodeId string) (*organization.Claim, error) {
	orgClaim, err := c.PostgresDB.GetOrganizationClaimByNodeId(ctx, userId, orgNodeId)
	if err != nil {
		return nil, fmt.Errorf("error getting orgClaim for user %d, workspace %s: %w", userId, orgNodeId, err)
	}
	return orgClaim, nil
}

func (c *ClaimsManager) GetTeamClaims(ctx context.Context, userId int64) ([]teamUser.Claim, error) {
	teamClaims, err := c.PostgresDB.GetTeamClaims(ctx, userId)
	if err != nil {
		return nil, err
	}

	return teamClaims, nil
}

func (c *ClaimsManager) GetTokenWorkspace() (TokenWorkspace, bool) {
	var workspace TokenWorkspace
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
	tokenOrg, tokenHasOrg := c.GetTokenWorkspace()
	if tokenHasOrg {
		return tokenOrg.Id
	}
	return currentUser.PreferredOrg
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
func getUser(ctx context.Context, q PennsievePgAPI, cognitoId string, isFromTokenPool bool) (*pgdbModels.User, error) {

	if isFromTokenPool {
		//var token pgdbModels.Token
		currentUser, err := q.GetUserByCognitoId(ctx, cognitoId)
		if err != nil {
			return nil, fmt.Errorf("unable to get user: %w", err)
		}
		return currentUser, nil

	} else {
		//var user pgdbModels.User
		currentUser, err := q.GetByCognitoId(ctx, cognitoId)
		if err != nil {
			return nil, fmt.Errorf("unable to get user: %w", err)
		}
		return currentUser, nil
	}
}

type TokenWorkspace struct {
	Id     int64
	NodeId string
}
