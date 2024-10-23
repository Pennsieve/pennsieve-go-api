package factory

import (
	"errors"
	"fmt"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/pennsieve/pennsieve-go-api/authorizer/helpers"
	log "github.com/sirupsen/logrus"
)

type AuthorizerFactory interface {
	Build([]string, map[string]string) (authorizers.Authorizer, error)
}

type CustomAuthorizerFactory struct{}

func NewCustomAuthorizerFactory() AuthorizerFactory {
	return &CustomAuthorizerFactory{}
}

func (f *CustomAuthorizerFactory) Build(identitySource []string, queryStringParameters map[string]string) (authorizers.Authorizer, error) {
	if !helpers.Matches(identitySource[0], `Bearer (?P<token>.*)`) {
		errorString := "token expected to be first identity source"
		log.Error(errorString)
		return nil, errors.New(errorString)
	}

	// immediately return the UserAuthorizer
	if len(identitySource) == 1 {
		return authorizers.NewUserAuthorizer(), nil
	}

	// where len(identitySource) > 1
	var hasManifestId bool
	manifest_id, hasManifestId := queryStringParameters["manifest_id"]
	if manifest_id == "" {
		hasManifestId = false
	}

	paramIdentitySource, err := helpers.DecodeIdentitySource(identitySource[1])
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf("could not decode identity source: %w", err)
	}

	switch {
	case helpers.Matches(paramIdentitySource, `N:dataset:`):
		return authorizers.NewDatasetAuthorizer(paramIdentitySource), nil
	case helpers.Matches(paramIdentitySource, `N:organization:`):
		return authorizers.NewWorkspaceAuthorizer(paramIdentitySource), nil
	case hasManifestId:
		return authorizers.NewManifestAuthorizer(paramIdentitySource), nil // will be deprecated
	default:
		return nil, errors.New("no suitable authorizer to process request")

	}
}
