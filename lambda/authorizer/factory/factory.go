package factory

import (
	"errors"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
)

type AuthorizerFactory interface {
	Build([]string) (authorizers.Authorizer, error)
}

type CustomAuthorizerFactory struct{}

func NewCustomAuthorizerFactory() AuthorizerFactory {
	return &CustomAuthorizerFactory{}
}

func (f *CustomAuthorizerFactory) Build(identitySource []string) (authorizers.Authorizer, error) {
	switch identitySource[0] {
	case "DatasetAuthorizer":
		return authorizers.NewDatasetAuthorizer(identitySource[2]), nil
	case "ManifestAuthorizer":
		return authorizers.NewManifestAuthorizer(identitySource[2]), nil // will be deprecated
	case "WorkspaceAuthorizer":
		return authorizers.NewWorkspaceAuthorizer(identitySource[2]), nil
	case "UserAuthorizer":
		return authorizers.NewUserAuthorizer(), nil
	default:
		return nil, errors.New("unsupported authorizer")
	}
}
