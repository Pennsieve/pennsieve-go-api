package factory

import (
	"errors"
	"fmt"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/pennsieve/pennsieve-go-api/authorizer/helpers"
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

	identitySourceMapper, err := mappers.NewIdentitySourceMapper(identitySource)
	if err != nil {
		return nil, err
	}
	mappedIdentitySource, err := identitySourceMapper.Create()
	if err != nil {
		return nil, err
	}

	// immediately return the UserAuthorizer
	if mappedIdentitySource.Other == nil {
		return authorizers.NewUserAuthorizer(), nil
	}

	otherIdentitySource := *mappedIdentitySource.Other

	if otherIdentitySource == queryStringParameters["dataset_id"] {
		paramIdentitySource, err := helpers.DecodeIdentitySource(otherIdentitySource)
		if err != nil {
			return nil, fmt.Errorf("could not decode dataset_id identity source: %w", err)
		}
		return authorizers.NewDatasetAuthorizer(paramIdentitySource), nil
	}

	if otherIdentitySource == queryStringParameters["organization_id"] {
		paramIdentitySource, err := helpers.DecodeIdentitySource(otherIdentitySource)
		if err != nil {
			return nil, fmt.Errorf("could not decode workspace_id identity source: %w", err)
		}
		return authorizers.NewWorkspaceAuthorizer(paramIdentitySource), nil
	}

	if otherIdentitySource == queryStringParameters["manifest_id"] {
		paramIdentitySource, err := helpers.DecodeIdentitySource(otherIdentitySource)
		if err != nil {
			return nil, fmt.Errorf("could not decode manifest_id identity source: %w", err)
		}
		return authorizers.NewManifestAuthorizer(paramIdentitySource), nil
	}
	return nil, errors.New("no suitable authorizer to process request")
}
