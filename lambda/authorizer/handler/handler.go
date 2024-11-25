package handler

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pennsieve/pennsieve-go-api/authorizer/helpers"
	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-api/authorizer/service"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/dydb"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
)

var keySet jwk.Set
var regionID string
var userPoolID string
var userClientID string
var tokenPoolID string
var tokenClientID string
var issuer string
var tokenIssuer string
var manifestTableName string

// init runs on cold start of lambda and gets jwt keysets from Cognito user pools.
func init() {
	regionID = os.Getenv("REGION")
	userPoolID = os.Getenv("USER_POOL")
	userClientID = os.Getenv("USER_CLIENT")
	tokenPoolID = os.Getenv("TOKEN_POOL")
	tokenClientID = os.Getenv("TOKEN_CLIENT")
	manifestTableName = os.Getenv("MANIFEST_TABLE")
	issuer = fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", regionID, userPoolID)
	tokenIssuer = fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", regionID, tokenPoolID)

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
		log.Error("Unable to fetch user pool Key Set", err)
	}

	// Get TokenPool keyset
	tokenJwksURL := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json", regionID, tokenPoolID)
	tokenKeySet, err := jwk.Fetch(context.Background(), tokenJwksURL)
	if err != nil {
		log.Error("Unable to fetch token pool Key Set", err)
	}

	// Add tokenKeySet keys to keySet, so we can decode from both user and token pool
	tokenKeys := tokenKeySet.Keys(context.Background())
	for tokenKeys.Next(context.Background()) {
		if err := keySet.AddKey(tokenKeys.Pair().Value.(jwk.Key)); err != nil {
			log.Error("Unable to add token pool keys to user pool keys", err)
		}
	}

}

// Handler runs in response to authorization event from the AWS API Gateway.
func Handler(ctx context.Context, event events.APIGatewayV2CustomAuthorizerV2Request) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {
	logger := log.WithFields(log.Fields{"Type": event.Type,
		"pathParameters":          event.PathParameters,
		"QueryStringParameters":   event.QueryStringParameters,
		"rawPath":                 event.RawPath,
		"requestContext.routeKey": event.RequestContext.RouteKey})

	// This Info call may be demoted to Debug or removed entirely in future
	logger.WithFields(log.Fields{
		"IdentitySource": event.IdentitySource,
		"Headers":        event.Headers}).Info("request parameters")

	jwtB64, err := helpers.GetJWT(event.Headers["authorization"])
	if err != nil {
		logger.Error(err)
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: false,
			Context:      nil,
		}, nil
	}

	// Validate and parse token, and return unauthorized if not valid
	token, err := validateCognitoJWT(jwtB64)
	if err != nil {
		logger.Error(err)
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: false,
			Context:      nil,
		}, nil
	}

	// Open Pennsieve DB Connection
	db, err := pgdb.ConnectRDS()
	postgresDB := pgdb.New(db)
	if err != nil {
		logger.Fatalln("unable to connect to RDS instance.")
	}
	defer db.Close()

	// Create a DynamoDB connection
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		logger.Fatalln("unable to connect to RDS instance.")
	}
	client := dynamodb.NewFromConfig(cfg)
	dynamoDB := dydb.New(client)

	// Get claims
	identityService := service.NewIdentitySourceService(event.IdentitySource, event.QueryStringParameters)
	authorizer, err := identityService.GetAuthorizer(ctx)
	if err != nil {
		logger.Error(err)
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{
			IsAuthorized: false,
			Context:      nil,
		}, nil
	}
	claimsManager := manager.NewClaimsManager(postgresDB, dynamoDB, token, tokenClientID, manifestTableName)
	authorizerMode := os.Getenv("AUTHORIZER_MODE")
	claims, err := authorizer.GenerateClaims(ctx, claimsManager, authorizerMode)
	if err != nil {
		logger.Error(err)
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

// validateCognitoJWT parses and validates the provided JWT from Cognito.
func validateCognitoJWT(jwtB64 []byte) (jwt.Token, error) {

	// Parse the JWT.
	token, err := jwt.Parse(jwtB64, jwt.WithKeySet(keySet))
	if err != nil {
		return nil, fmt.Errorf("error parsing JWT: %w", err)
	}

	if token.Issuer() != issuer && token.Issuer() != tokenIssuer {
		return nil, fmt.Errorf("AUTHORIZER_FAILURE: Issuer in token does not match Pennsieve token issuers: %s", token.Issuer())
	}

	clientIdClaim, hasKey := token.Get("client_id")
	if !hasKey || (clientIdClaim != userClientID && clientIdClaim != tokenClientID) {
		detail := clientIdClaim
		if !hasKey {
			detail = "client_id missing"
		}
		return nil, fmt.Errorf("unauthorized: audience in token does not match: %s", detail)
	}

	if token.Expiration().Unix() < time.Now().Unix() {
		return nil, errors.New("unauthorized: token expired")
	}

	tokenUseClaim, hasKey := token.Get("token_use")
	if !hasKey || tokenUseClaim != "access" {
		detail := tokenUseClaim
		if !hasKey {
			detail = "token_use missing"
		}
		return nil, fmt.Errorf("unauthorized: Incorrect TokenUse Claim: %s", detail)
	}

	return token, err
}
