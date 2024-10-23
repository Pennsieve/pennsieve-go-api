package factory

import (
	"errors"

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

	paramIdentifier, err := helpers.DecodeIdentifier(identitySource[1])
	if err != nil {
		errorString := "could not decode identity source"
		log.Error(errorString)
		return nil, errors.New(errorString)
	}

	switch {
	case helpers.Matches(paramIdentifier, `N:dataset:`):
		return authorizers.NewDatasetAuthorizer(paramIdentifier), nil
	case helpers.Matches(paramIdentifier, `N:organization:`):
		return authorizers.NewWorkspaceAuthorizer(paramIdentifier), nil
	case hasManifestId:
		return authorizers.NewManifestAuthorizer(paramIdentifier), nil // will be deprecated
	default:
		return nil, errors.New("no suitable authorizer to process request")

	}
}
