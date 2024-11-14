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
	var hasManifestId bool
	manifest_id, hasManifestId := queryStringParameters["manifest_id"]
	if manifest_id == "" {
		hasManifestId = false
	}

	identitySourceMapper := NewIdentitySourceMapper(identitySource, hasManifestId)
	auxiliaryIdentitySource := identitySourceMapper.Create()

	// immediately return the UserAuthorizer
	_, ok := auxiliaryIdentitySource["token"]
	if ok && len(auxiliaryIdentitySource) == 1 {
		return authorizers.NewUserAuthorizer(), nil
	}

	if len(auxiliaryIdentitySource) > 1 {
		datasetID, ok := auxiliaryIdentitySource["dataset_id"]
		if ok {
			paramIdentitySource, err := helpers.DecodeIdentitySource(datasetID)
			if err != nil {
				log.Error(err)
				return nil, fmt.Errorf("could not decode identity source: %w", err)
			}
			return authorizers.NewDatasetAuthorizer(paramIdentitySource), nil
		}

		workspaceID, ok := auxiliaryIdentitySource["workspace_id"]
		if ok {
			paramIdentitySource, err := helpers.DecodeIdentitySource(workspaceID)
			if err != nil {
				log.Error(err)
				return nil, fmt.Errorf("could not decode identity source: %w", err)
			}
			return authorizers.NewWorkspaceAuthorizer(paramIdentitySource), nil
		}

		manifestID, ok := auxiliaryIdentitySource["manifest_id"]
		if ok {
			paramIdentitySource, err := helpers.DecodeIdentitySource(manifestID)
			if err != nil {
				log.Error(err)
				return nil, fmt.Errorf("could not decode identity source: %w", err)
			}
			return authorizers.NewManifestAuthorizer(paramIdentitySource), nil
		}
	}
	return nil, errors.New("no suitable authorizer to process request")
}
