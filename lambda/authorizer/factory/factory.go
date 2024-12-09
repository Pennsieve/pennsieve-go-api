package factory

import (
	"errors"
	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/pennsieve/pennsieve-go-api/authorizer/mappers"
)

type AuthorizerFactory interface {
	Build([]string, map[string]string) (authorizers.Authorizer, error)
}

type CustomAuthorizerFactory struct{}

func NewCustomAuthorizerFactory() AuthorizerFactory {
	return &CustomAuthorizerFactory{}
}

func (f *CustomAuthorizerFactory) Build(identitySource []string, queryStringParameters map[string]string) (authorizers.Authorizer, error) {

	mappedIdentitySource, err := mappers.NewIdentitySourceMapper(identitySource).Create()
	if err != nil {
		return nil, err
	}

	// immediately return the UserAuthorizer
	if mappedIdentitySource.Other == nil {
		return authorizers.NewUserAuthorizer(), nil
	}

	otherIdentitySource := *mappedIdentitySource.Other

	if otherIdentitySource == queryStringParameters["dataset_id"] {
		return authorizers.NewDatasetAuthorizer(otherIdentitySource), nil
	}

	if otherIdentitySource == queryStringParameters["organization_id"] {
		return authorizers.NewWorkspaceAuthorizer(otherIdentitySource), nil
	}

	if otherIdentitySource == queryStringParameters["manifest_id"] {
		return authorizers.NewManifestAuthorizer(otherIdentitySource), nil
	}
	return nil, errors.New("no suitable authorizer to process request")
}
