package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pennsieve/pennsieve-go-api/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-api/pkg/core"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/dbTable"
	"github.com/pennsieve/pennsieve-go-api/pkg/models/user"
	log "github.com/sirupsen/logrus"
	"os"
	"regexp"
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

	// Get Identity Sources
	// If single identity source, then no dataset claim should be generated.
	r := regexp.MustCompile(`Bearer (?P<token>.*)`)
	tokenParts := r.FindStringSubmatch(event.Headers["authorization"])
	jwtB64 := []byte(tokenParts[r.SubexpIndex("token")])

	datasetNodeId, hasDatasetId := event.PathParameters["dataset_id"]
	if !hasDatasetId {
		datasetNodeId, hasDatasetId = event.QueryStringParameters["dataset_id"]
	}
	if hasDatasetId && len(event.IdentitySource) < 2 {
		log.Fatalln("Request cannot have dataset_id as query-param or path-param with the used authorizer.")
	}

	manifestId, hasManifestId := event.QueryStringParameters["manifest_id"]
	if hasManifestId && len(event.IdentitySource) < 2 {
		log.Fatalln("Request cannot have manifest_id as query-param with the used authorizer.")
	}

	if hasDatasetId && hasManifestId {
		log.Fatalln("Request cannot have both manifest_id and dataset_id as query-params with the used authorizer.")
	}

	// Get Dataset associated with the requested manifest
	if hasManifestId {
		cfg, err := config.LoadDefaultConfig(context.Background())
		if err != nil {
			panic("unable to load SDK config, " + err.Error())
		}

		// Create an Amazon DynamoDB client.
		client := dynamodb.NewFromConfig(cfg)
		table := os.Getenv("MANIFEST_TABLE")

		manifest, err := dbTable.GetFromManifest(client, table, manifestId)
		if err != nil {
			log.Fatalln("Manifest could not be found: ", err)
		}

		datasetNodeId = manifest.DatasetNodeId
		hasDatasetId = true
	}

	// Validate and parse token, and return unauthorized if not valid
	token, err := validateCognitoJWT(jwtB64)
	if err != nil {
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("Unauthorized")
	}

	// Open Pennsieve DB Connection
	db, err := core.ConnectRDS()
	if err != nil {
		log.Fatalln("Unable to connect to RDS instance.")
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
	currentUser, err := getUser(db, cognitoUserName.(string), isFromTokenPool)
	if err != nil {
		log.Fatalln("Unable to get User from Cognito Username")
	}

	// Get Active Org
	orgInt := currentUser.PreferredOrg
	jwtOrg, hasKey := token.Get("custom:organization_id")
	if hasKey {
		orgInt = jwtOrg.(int64)
	}

	// Get ORG Claim
	orgClaim, err := authorizer.GetOrganizationClaim(db, currentUser.Id, orgInt)
	if err != nil {
		log.Error("Unable to get Organization Role")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("Unauthorized") // Return 401: Unauthenticated
	}

	// Get DATASET Claim
	var datasetClaim *dataset.Claim
	if hasDatasetId {
		datasetClaim, err = authorizer.GetDatasetClaim(db, currentUser, datasetNodeId, orgInt)
		if err != nil {
			log.Error("Unable to get Dataset Role")
			return events.APIGatewayV2CustomAuthorizerSimpleResponse{
				IsAuthorized: false,
				Context:      nil,
			}, nil // Return 403: Forbidden
		}

		// If user has no role on provided dataset --> return
		if datasetClaim.Role == dataset.None {
			log.Error("User has no access to dataset")
			return events.APIGatewayV2CustomAuthorizerSimpleResponse{
				IsAuthorized: false,
				Context:      nil,
			}, nil // Return 403: Forbidden
		}

	} else {
		datasetClaim = nil
	}

	userClaim := user.Claim{
		Id:           currentUser.Id,
		NodeId:       currentUser.NodeId,
		IsSuperAdmin: currentUser.IsSuperAdmin,
	}

	// Bundle Claims
	claims := map[string]interface{}{
		"user_claim":    userClaim,
		"org_claim":     orgClaim,
		"dataset_claim": datasetClaim,
	}

	return events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: true,
		Context:      claims,
	}, nil

}

// getUser returns a Pennsieve user from a cognito ID.
func getUser(db core.PostgresAPI, cognitoId string, isFromTokenPool bool) (*dbTable.User, error) {

	if isFromTokenPool {
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
