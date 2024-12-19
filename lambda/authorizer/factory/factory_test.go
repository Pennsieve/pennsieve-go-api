package factory_test

import (
	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/factory"
	"github.com/stretchr/testify/assert"
)

func TestFactory(t *testing.T) {
	authFactory := factory.NewCustomAuthorizerFactory()

	//ids
	objectId := "someObjectId"
	object2Id := "someOtherObjectId"

	//user token header
	authHeaderValue := "Bearer eyJra.some.random.string"

	// query params
	withManifestId := map[string]string{"manifest_id": objectId, "someOtherParam": "someOtherValue"}
	withDatasetId := map[string]string{"dataset_id": objectId, "someOtherParam": "someOtherValue"}
	withWorkspaceId := map[string]string{"organization_id": objectId, "someOtherParam": "someOtherValue"}
	withoutIdQueryParams := map[string]string{"someOtherParam": "someOtherValue"}
	withDatasetAndManifestIds := map[string]string{"dataset_id": objectId, "manifest_id": object2Id, "someOtherParam": "someOtherValue"}
	withDatasetAndWorkspaceIds := map[string]string{"dataset_id": objectId, "organization_id": object2Id, "someOtherParam": "someOtherValue"}

	// identity sources
	objectIdentitySource := []string{authHeaderValue, objectId}
	objectIdentitySourceFlippedOrder := []string{objectId, authHeaderValue}
	object2IdentitySource := []string{object2Id, authHeaderValue}
	userIdentitySource := []string{authHeaderValue}

	// expected authorizer types
	var userAuthorizerType *authorizers.UserAuthorizer
	var datasetAuthorizerType *authorizers.DatasetAuthorizer
	var workspaceAuthorizerType *authorizers.WorkspaceAuthorizer
	var manifestAuthorizerType *authorizers.ManifestAuthorizer

	// happy path tests
	for scenario, params := range map[string]struct {
		idSource               []string
		queryParams            map[string]string
		expectedAuthorizerType authorizers.Authorizer
	}{
		"user authorizer": {userIdentitySource, withoutIdQueryParams, userAuthorizerType},
		"user authorizer with id related query params":                                 {userIdentitySource, withDatasetId, userAuthorizerType},
		"dataset authorizer":                                                           {objectIdentitySource, withDatasetId, datasetAuthorizerType},
		"dataset authorizer, flipped identity source":                                  {objectIdentitySourceFlippedOrder, withDatasetId, datasetAuthorizerType},
		"manifest authorizer":                                                          {objectIdentitySource, withManifestId, manifestAuthorizerType},
		"manifest authorizer, flipped identity source":                                 {objectIdentitySourceFlippedOrder, withManifestId, manifestAuthorizerType},
		"workspace authorizer":                                                         {objectIdentitySource, withWorkspaceId, workspaceAuthorizerType},
		"workspace authorizer, flipped identity source":                                {objectIdentitySourceFlippedOrder, withWorkspaceId, workspaceAuthorizerType},
		"user supplies both manifest and dataset id to manifest authorizer endpoint":   {object2IdentitySource, withDatasetAndManifestIds, manifestAuthorizerType},
		"user supplies both manifest and dataset id to dataset authorizer endpoint":    {objectIdentitySource, withDatasetAndManifestIds, datasetAuthorizerType},
		"user supplies both workspace and dataset id to workspace authorizer endpoint": {object2IdentitySource, withDatasetAndWorkspaceIds, workspaceAuthorizerType},
		"user supplies both workspace and dataset id to dataset authorizer endpoint":   {objectIdentitySource, withDatasetAndWorkspaceIds, datasetAuthorizerType},
	} {
		t.Run(scenario, func(t *testing.T) {
			authorizer, err := authFactory.Build(params.idSource, params.queryParams)
			require.NoError(t, err)
			assert.IsType(t, params.expectedAuthorizerType, authorizer)
		})
	}

	missingUserTokenIdentitySource := []string{objectId}
	userIdentitySourceMissingBearer := []string{"eyJra.some.random.string"}
	userIdentitySourceOnlyBearer := []string{"Bearer"}

	idSourceEmptyObjectId := []string{authHeaderValue, ""}
	queryParamsEmptyObjectId := map[string]string{"dataset_id": ""}

	// error tests
	for scenario, params := range map[string]struct {
		idSource          []string
		queryParams       map[string]string
		expectedErrorText string
	}{
		"missing user token":           {missingUserTokenIdentitySource, withWorkspaceId, "no valid user token found"},
		"user token missing 'Bearer'":  {userIdentitySourceMissingBearer, withoutIdQueryParams, "no valid user token found"},
		"user token only has 'Bearer'": {userIdentitySourceOnlyBearer, withoutIdQueryParams, "no valid user token found"},
		"empty object id":              {idSourceEmptyObjectId, queryParamsEmptyObjectId, "invalid non-token identity source found"},
	} {
		t.Run(scenario, func(t *testing.T) {
			authorizer, err := authFactory.Build(params.idSource, params.queryParams)
			assert.Equal(t, nil, authorizer)
			assert.ErrorContains(t, err, params.expectedErrorText)
		})
	}
}
