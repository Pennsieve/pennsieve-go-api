package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
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
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("INVALID_API_KEY")
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
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("Unauthorized")
	}

	if token.Expiration().Unix() < time.Now().Unix() {
		fmt.Println("Token expired.")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("Unauthorized")
	}

	tokenUseClaim, hasKey := token.Get("token_use")
	if hasKey != true || tokenUseClaim != "accjhess" {
		fmt.Println("Incorrect TokenUse Claim")
		return events.APIGatewayV2CustomAuthorizerSimpleResponse{}, errors.New("Unauthorized")
	}

	/*
		At this point, the user is authenticated, now we check the org, dataset and generate claims.
	*/

	log.Println("The token is valid.")

	str, _ := json.Marshal(event)
	str2 := string(str)
	fmt.Println(str2)

	return events.APIGatewayV2CustomAuthorizerSimpleResponse{
		IsAuthorized: true,
		Context:      nil,
	}, nil

}
