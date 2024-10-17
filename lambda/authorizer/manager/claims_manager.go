package manager

import (
	"context"
	"errors"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	pgdbQueries "github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
)

type IdentityManager interface {
	// GetUserClaim(context.Context) user.Claim
	GetDatasetClaim(context.Context, *pgdbModels.User, string, int64) (*dataset.Claim, error)
	GetCurrentUser(context.Context) (*pgdbModels.User, error)
	GetToken() jwt.Token
	GetQueryHandle() *pgdbQueries.Queries
	GetTokenClientID() string
	// GetOrgClaim(context.Context) *organization.Claim
	// GetTeamsClaim(context.Context) []teamUser.Claim
}

type ClaimsManager struct {
	QueryHandle   *pgdbQueries.Queries
	Token         jwt.Token
	TokenClientID string
}

func NewClaimsManager(queryHandle *pgdbQueries.Queries, token jwt.Token, tokenClientID string) IdentityManager {
	return &ClaimsManager{queryHandle, token, tokenClientID}
}

func (c *ClaimsManager) GetDatasetClaim(ctx context.Context, currentUser *pgdbModels.User, datasetId string, orgInt int64) (*dataset.Claim, error) {
	var datasetClaim *dataset.Claim
	datasetClaim, err := c.QueryHandle.GetDatasetClaim(ctx, currentUser, datasetId, orgInt)
	if err != nil {
		log.Error("unable to get Dataset Role")
		return nil, err
	}

	return datasetClaim, nil
}

func (c *ClaimsManager) GetToken() jwt.Token {
	return c.Token
}

func (c *ClaimsManager) GetTokenClientID() string {
	return c.TokenClientID
}

func (c *ClaimsManager) GetQueryHandle() *pgdbQueries.Queries {
	return c.QueryHandle
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
	return getUser(ctx, c.QueryHandle, cognitoUserName.(string), isFromTokenPool)
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
