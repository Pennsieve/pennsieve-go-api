package service

import (
	"context"
	"errors"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/pennsieve/pennsieve-go-api/authorizer/mappers"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	pgdbQueries "github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
)

type IdentityService interface {
	GetCurrentUser(context.Context, string) (*pgdbModels.User, error)
	GetClaims(context.Context) (map[string]interface{}, error)
	GetAuthorizer(context.Context, *pgdbModels.User) (authorizers.Authorizer, error)
}

type IdentitySourceService struct {
	IdentitySource []string
	Token          jwt.Token
	QueryHandle    *pgdbQueries.Queries
	TokenClientID  string
}

func NewIdentitySourceService(IdentitySource []string, token jwt.Token, queryHandle *pgdbQueries.Queries, tokenClientID string) IdentityService {
	return &IdentitySourceService{IdentitySource, token, queryHandle, tokenClientID}
}

func (i *IdentitySourceService) GetCurrentUser(ctx context.Context, tokenClientID string) (*pgdbModels.User, error) {
	// Get Cognito User ID
	cognitoUserName, hasKey := i.Token.Get("username")
	if !hasKey {
		return nil, errors.New("Unauthorized")
	}

	// Get Pennsieve User from User Table, or Token Table
	clientIdClaim, _ := i.Token.Get("client_id") // Key is present or method would have returned before.
	isFromTokenPool := clientIdClaim == tokenClientID
	return getUser(ctx, i.QueryHandle, cognitoUserName.(string), isFromTokenPool)

}

func (i *IdentitySourceService) GetClaims(ctx context.Context) (map[string]interface{}, error) {
	currentUser, err := i.GetCurrentUser(ctx, i.TokenClientID)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	authorizer, err := i.GetAuthorizer(ctx, currentUser)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	return authorizer.GenerateClaims(ctx)
}

func (i *IdentitySourceService) GetAuthorizer(ctx context.Context, currentUser *pgdbModels.User) (authorizers.Authorizer, error) {
	authFactory := mappers.NewCustomAuthorizerFactory(currentUser, i.QueryHandle, i.Token)
	return mappers.IdentitySourceToAuthorizer(i.IdentitySource, authFactory)
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
