package handler

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pennsieve/pennsieve-go-api/models"
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

func init() {
	regionID = os.Getenv("REGION")
	userPoolID = os.Getenv("USER_POOL")
	userClientID = os.Getenv("USER_CLIENT")
	tokenPoolID = os.Getenv("TOKEN_POOL")
	tokenClientID = os.Getenv("TOKEN_CLIENT")

	// Get UserPool keyset
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

func Handler(ctx context.Context, event events.APIGatewayV2CustomAuthorizerV2Request) (events.APIGatewayV2CustomAuthorizerSimpleResponse, error) {

	// See the AWS docs here:
	// https://docs.aws.amazon.com/cognito/latest/developerguide/amazon-cognito-user-pools-using-tokens-verifying-a-jwt.html
	jwtB64 := []byte(event.Headers["authorization"])

	// Parse the JWT.
	token, err := jwt.Parse(jwtB64, jwt.WithKeySet(keySet))
	if err != nil {
		fmt.Printf("Failed to parse the JWT.\nError:%s\n\n", err.Error())
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("unauthorized")
	}

	issuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", regionID, userPoolID)
	tokenIssuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", regionID, tokenPoolID)
	if token.Issuer() != issuer && token.Issuer() != tokenIssuer {
		fmt.Println("Issuer in token does not match.")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{},
			errors.New("AUTHORIZER_FAILURE: Issuer in token does not match Pennsieve token issuers")
	}

	clientIdClaim, hasKey := token.Get("client_id")
	if hasKey != true || (clientIdClaim != userClientID && clientIdClaim != tokenClientID) {
		fmt.Println("Audience in token does not match.")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("unauthorized")
	}

	if token.Expiration().Unix() < time.Now().Unix() {
		fmt.Println("Token expired.")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("unauthorized")
	}

	tokenUseClaim, hasKey := token.Get("token_use")
	if hasKey != true || tokenUseClaim != "access" {
		fmt.Println("Incorrect TokenUse Claim")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("unauthorized")
	}

	/*
		At this point, the user is authenticated, now we check the org, dataset and generate claims.
	*/

	log.Println("The token is valid.")

	db, err := core.ConnectRDS()
	defer db.Close()

	cognitoUserName, hasKey := token.Get("username")
	if hasKey != true {
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("unauthorized")
	}

	fmt.Println("Cognito-user: ", cognitoUserName.(string))
	var user models.User
	currentUser, err := user.GetByCognitoId(db, cognitoUserName.(string))
	if err != nil {
		log.Fatalln("Unable to get user:", err)
	}

	fmt.Println("GETTING USER")
	fmt.Println(currentUser)

	return events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: true,
		Context:      nil,
	}, nil

}
