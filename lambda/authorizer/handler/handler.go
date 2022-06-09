package handler

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pennsieve/pennsieve-go-api/models/dbTable"
	"github.com/pennsieve/pennsieve-go-api/models/organization"
	"github.com/pennsieve/pennsieve-go-api/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-api/pkg/core"
	"log"
	"os"
	"time"
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

	// Get UserPool keyset
	// https://docs.aws.amazon.com/cognito/latest/developerguide/amazon-cognito-user-pools-using-tokens-verifying-a-jwt.html
	userJwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", regionID, userPoolID)
	keySet, err = jwk.Fetch(context.Background(), userJwksURL)
	if err != nil {
		fmt.Println("Unable to fetch Key Set")
	}

	// Get TokenPool keyset
	tokenJwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", regionID, tokenPoolID)
	tokenKeySet, err := jwk.Fetch(context.Background(), tokenJwksURL)
	if err != nil {
		fmt.Println("Unable to fetch Key Set")
	}

	// Add tokenKeySet keys to keySet, so we can decode from both user and token pool
	tokenKeys := tokenKeySet.Keys(context.Background())
	for tokenKeys.Next(context.Background()) {
		keySet.AddKey(tokenKeys.Pair().Value.(jwk.Key))
	}

}

// Handler runs in response to authorization event from the AWS API Gateway.
func Handler(ctx context.Context, event events.APIGatewayV2CustomAuthorizerV2Request) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {

	jwtB64 := []byte(event.Headers["authorization"])
	datasetNodeId, hasKey := event.QueryStringParameters["dataset_id"]

	// Validate and parse token, and return unauthorized if not valid
	token, err := validateCognitoJWT(jwtB64)
	if err != nil {
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("unauthorized")
	}

	db, err := core.ConnectRDS()
	if err != nil {
		log.Fatalln("Unable to connect to RDS instance.")
	}
	defer db.Close()

	// Get Cognito User ID
	cognitoUserName, hasKey := token.Get("username")
	if hasKey != true {
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("unauthorized")
	}

	// Get Pennsieve User from User Table, or Token Table
	clientIdClaim, _ := token.Get("client_id") // Key is present or method would have returned before.
	isFromTokenPool := clientIdClaim == tokenClientID

	currentUser, err := getUser(db, cognitoUserName.(string), isFromTokenPool)
	if err != nil {
		log.Fatalln("Unable to get User from Cognito Username")
	}

	// Get Active Org
	orgInt := currentUser.PreferredOrg
	jwtOrg, hasKey := token.Get("custom:organization_id")
	if hasKey {
		fmt.Println("Using Cognito Org ID from JWT")
		orgInt = jwtOrg.(int64)
	}

	// Get ORG Claim
	orgClaim, err := organization.GetOrganizationClaim(db, currentUser.Id, orgInt)
	if err != nil {
		fmt.Println("Unable to get Organization Role")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("unauthorized")
	}

	// Get DATASET Claim
	datasetClaim, err := authorizer.GetDatasetClaim(db, currentUser, datasetNodeId, orgInt)
	if err != nil {
		fmt.Println("Unable to get Dataset Role")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("unauthorized")
	}

	// Bundle Claims
	claims := map[string]interface{}{
		"organization_id": orgInt,
		"user_id":         currentUser.Id,
		"is_super_admin":  currentUser.IsSuperAdmin,
		"org_claim":       orgClaim,
		"dataset_claim":   datasetClaim,
	}

	fmt.Println(claims)

	return events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: true,
		Context:      claims,
	}, nil

}

// getUser returns a Pennsieve user from a cognito ID.
func getUser(db *sql.DB, cognitoId string, isFromTokenPool bool) (*dbTable.User, error) {

	if isFromTokenPool {
		fmt.Println("Getting User from Token Pool")
		var token dbTable.Token
		currentUser, err := token.GetUserByCognitoId(db, cognitoId)
		if err != nil {
			log.Fatalln("Unable to get user:", err)
		}
		return currentUser, nil

	} else {
		var user dbTable.User
		currentUser, err := user.GetByCognitoId(db, cognitoId)
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
		fmt.Printf("Failed to parse the JWT.\nError:%s\n\n", err.Error())
		return nil, errors.New("unauthorized")
	}

	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", regionID, userPoolID)
	tokenIssuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", regionID, tokenPoolID)
	if token.Issuer() != issuer && token.Issuer() != tokenIssuer {
		fmt.Println("Issuer in token does not match.")
		return nil, errors.New("AUTHORIZER_FAILURE: Issuer in token does not match Pennsieve token issuers")
	}

	clientIdClaim, hasKey := token.Get("client_id")
	if hasKey != true || (clientIdClaim != userClientID && clientIdClaim != tokenClientID) {
		fmt.Println("Audience in token does not match.")
		return nil, errors.New("unauthorized")
	}

	if token.Expiration().Unix() < time.Now().Unix() {
		fmt.Println("Token expired.")
		return nil, errors.New("unauthorized")
	}

	tokenUseClaim, hasKey := token.Get("token_use")
	if hasKey != true || tokenUseClaim != "access" {
		fmt.Println("Incorrect TokenUse Claim")
		return nil, errors.New("unauthorized")
	}

	return token, err
}
