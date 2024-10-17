package factory

import (
	"errors"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/pennsieve/pennsieve-go-api/authorizer/helpers"
	log "github.com/sirupsen/logrus"
)

type AuthorizerFactory interface {
	Build([]string) (authorizers.Authorizer, error)
}

type CustomAuthorizerFactory struct{}

func NewCustomAuthorizerFactory() AuthorizerFactory {
	return &CustomAuthorizerFactory{}
}

func (f *CustomAuthorizerFactory) Build(identitySource []string) (authorizers.Authorizer, error) {
	if !helpers.Matches(identitySource[0], `Bearer (?P<token>.*)`) {
		errorString := "token expected to be first identity source"
		log.Error(errorString)
		return nil, errors.New(errorString)
	}

	switch {
	case len(identitySource) > 1 && helpers.Matches(identitySource[1], `N:dataset:`):
		return authorizers.NewDatasetAuthorizer(identitySource[1]), nil
	case len(identitySource) > 1 && helpers.Matches(identitySource[1], `N:manifest:`):
		return authorizers.NewManifestAuthorizer(identitySource[1]), nil // will be deprecated
	case len(identitySource) > 1 && helpers.Matches(identitySource[1], `N:organization:`):
		return authorizers.NewWorkspaceAuthorizer(identitySource[1]), nil
	default:
		return authorizers.NewUserAuthorizer(), nil
	}
}
