package handler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pennsieve/pennsieve-go-api/authorizer/mappers"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
)

var err error
var keySet jwk.Set
var regionID string
var userPoolID string
var userClientID string
var tokenPoolID string
var tokenClientID string

// init runs on cold start of lambda and gets jwt keysets from Cognito user pools.
func init() {
	regionID = os.Getenv("REGION")
	userPoolID = os.Getenv("USER_POOL")
	userClientID = os.Getenv("USER_CLIENT")
	tokenPoolID = os.Getenv("TOKEN_POOL")
	tokenClientID = os.Getenv("TOKEN_CLIENT")

	log.SetFormatter(&log.JSONFormatter{})
	ll, err := log.ParseLevel(os.Getenv("LOG_LEVEL"))
	if err != nil {
		log.SetLevel(log.InfoLevel)
	} else {
		log.SetLevel(ll)
	}

	// Get UserPool keyset
	// https://docs.aws.amazon.com/cognito/latest/developerguide/amazon-cognito-user-pools-using-tokens-verifying-a-jwt.html
	userJwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", regionID, userPoolID)
	keySet, err = jwk.Fetch(context.Background(), userJwksURL)
	if err != nil {
		log.Error("Unable to fetch Key Set")
	}

	// Get TokenPool keyset
	tokenJwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", regionID, tokenPoolID)
	tokenKeySet, err := jwk.Fetch(context.Background(), tokenJwksURL)
	if err != nil {
		log.Error("Unable to fetch Key Set")
	}

	// Add tokenKeySet keys to keySet, so we can decode from both user and token pool
	tokenKeys := tokenKeySet.Keys(context.Background())
	for tokenKeys.Next(context.Background()) {
		keySet.AddKey(tokenKeys.Pair().Value.(jwk.Key))
	}

}

// Handler runs in response to authorization event from the AWS API Gateway.
func Handler(ctx context.Context, event events.APIGatewayV2CustomAuthorizerV2Request) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {

	log.Info("request parameters",
		"Type", event.Type,
		"IdentitySource", event.IdentitySource,
		"pathParameters", event.PathParameters,
		"QueryStringParameters", event.QueryStringParameters,
		"rawPath", event.RawPath,
		"Headers", event.Headers,
		"requestContext.routeKey", event.RequestContext.RouteKey,
		"event.RequestContext.Authorizer", event.RequestContext.Authorizer)

	r := regexp.MustCompile(`Bearer (?P<token>.*)`)
	tokenParts := r.FindStringSubmatch(event.Headers["authorization"])
	jwtB64 := []byte(tokenParts[r.SubexpIndex("token")])

	// Validate and parse token, and return unauthorized if not valid
	token, err := validateCognitoJWT(jwtB64)
	if err != nil {
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("Unauthorized")
	}

	// Open Pennsieve DB Connection
	db, err := pgdb.ConnectRDS()
	postgresDB := pgdb.New(db)

	if err != nil {
		log.Fatalln("unable to connect to RDS instance.")
	}
	defer db.Close()

	// Get Cognito User ID
	cognitoUserName, hasKey := token.Get("username")
	if hasKey != true {
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("Unauthorized")
	}

	// Get Pennsieve User from User Table, or Token Table
	clientIdClaim, _ := token.Get("client_id") // Key is present or method would have returned before.
	isFromTokenPool := clientIdClaim == tokenClientID
	currentUser, err := getUser(ctx, postgresDB, cognitoUserName.(string), isFromTokenPool)
	if err != nil {
		log.Fatalln("unable to get User from Cognito Username")
	}

	// Get authorizer
	authorizer := mappers.IdentitySourceToAuthorizer(event.IdentitySource, currentUser, postgresDB, token)
	claims, err := authorizer.GenerateClaims(ctx)
	if err != nil {
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: false,
			Context:      nil,
		}, nil
	}

	return events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: true,
		Context:      claims,
	}, nil

}

// getUser returns a Pennsieve user from a cognito ID.
func getUser(ctx context.Context, q *pgdb.Queries, cognitoId string, isFromTokenPool bool) (*pgdbModels.User, error) {

	if isFromTokenPool {
		//var token pgdbModels.Token
		currentUser, err := q.GetUserByCognitoId(ctx, cognitoId)
		if err != nil {
			log.Fatalln("Unable to get user:", err)
		}
		return currentUser, nil

	} else {
		//var user pgdbModels.User
		currentUser, err := q.GetByCognitoId(ctx, cognitoId)
		if err != nil {
			log.Fatalln("Unable to get user:", err)
		}
		return currentUser, nil
	}

}

// validateCognitoJWT parses and validates the provided JWT from Cognito.
func validateCognitoJWT(jwtB64 []byte) (jwt.Token, error) {

	// Parse the JWT.
	token, err := jwt.Parse(jwtB64, jwt.WithKeySet(keySet))
	if err != nil {
		log.Debug(fmt.Sprintf("Failed to parse the JWT.\nError:%s\n\n", err.Error()))
		return nil, errors.New("unauthorized")
	}

	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", regionID, userPoolID)
	tokenIssuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", regionID, tokenPoolID)
	if token.Issuer() != issuer && token.Issuer() != tokenIssuer {
		log.Debug("Issuer in token does not match.")
		return nil, errors.New("AUTHORIZER_FAILURE: Issuer in token does not match Pennsieve token issuers")
	}

	clientIdClaim, hasKey := token.Get("client_id")
	if hasKey != true || (clientIdClaim != userClientID && clientIdClaim != tokenClientID) {
		log.Debug("Audience in token does not match.")
		return nil, errors.New("unauthorized")
	}

	if token.Expiration().Unix() < time.Now().Unix() {
		log.Debug("Token expired.")
		return nil, errors.New("unauthorized")
	}

	tokenUseClaim, hasKey := token.Get("token_use")
	if hasKey != true || tokenUseClaim != "access" {
		log.Debug("Incorrect TokenUse Claim")
		return nil, errors.New("unauthorized")
	}

	return token, err
}
