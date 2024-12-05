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
	manifestId := "someManifestId"
	datasetId := "N:dataset:some-uuid"
	workspaceId := "N:organization:some-uuid"

	//user token header
	authHeaderValue := "Bearer eyJra.some.random.string"

	// query params
	withManifestId := map[string]string{"manifest_id": manifestId, "someOtherParam": "someOtherValue"}
	withDatasetId := map[string]string{"dataset_id": datasetId, "someOtherParam": "someOtherValue"}
	withWorkspaceId := map[string]string{"organization_id": workspaceId, "someOtherParam": "someOtherValue"}
	withoutIdQueryParams := map[string]string{"someOtherParam": "someOtherValue"}
	withDatasetAndManifestIds := map[string]string{"dataset_id": datasetId, "manifest_id": manifestId, "someOtherParam": "someOtherValue"}
	withDatasetAndWorkspaceIds := map[string]string{"dataset_id": datasetId, "organization_id": workspaceId, "someOtherParam": "someOtherValue"}

	// identity sources
	manifestIdentitySource := []string{authHeaderValue, manifestId}
	manifestIdentitySourceFlippedOrder := []string{manifestId, authHeaderValue}
	datasetIdentitySource := []string{authHeaderValue, datasetId}
	datasetIdentitySourceFlippedOrder := []string{datasetId, authHeaderValue}
	workspaceIdentitySource := []string{authHeaderValue, workspaceId}
	workspaceIdentitySourceFlippedOrder := []string{workspaceId, authHeaderValue}
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
		"dataset authorizer":                                                           {datasetIdentitySource, withDatasetId, datasetAuthorizerType},
		"dataset authorizer, flipped identity source":                                  {datasetIdentitySourceFlippedOrder, withDatasetId, datasetAuthorizerType},
		"manifest authorizer":                                                          {manifestIdentitySource, withManifestId, manifestAuthorizerType},
		"manifest authorizer, flipped identity source":                                 {manifestIdentitySourceFlippedOrder, withManifestId, manifestAuthorizerType},
		"workspace authorizer":                                                         {workspaceIdentitySource, withWorkspaceId, workspaceAuthorizerType},
		"workspace authorizer, flipped identity source":                                {workspaceIdentitySourceFlippedOrder, withWorkspaceId, workspaceAuthorizerType},
		"user supplies both manifest and dataset id to manifest authorizer endpoint":   {manifestIdentitySource, withDatasetAndManifestIds, manifestAuthorizerType},
		"user supplies both manifest and dataset id to dataset authorizer endpoint":    {datasetIdentitySource, withDatasetAndManifestIds, datasetAuthorizerType},
		"user supplies both workspace and dataset id to workspace authorizer endpoint": {workspaceIdentitySource, withDatasetAndWorkspaceIds, workspaceAuthorizerType},
		"user supplies both workspace and dataset id to dataset authorizer endpoint":   {datasetIdentitySource, withDatasetAndWorkspaceIds, datasetAuthorizerType},
	} {
		t.Run(scenario, func(t *testing.T) {
			authorizer, err := authFactory.Build(params.idSource, params.queryParams)
			require.NoError(t, err)
			assert.IsType(t, params.expectedAuthorizerType, authorizer)
		})
	}

	missingUserTokenIdentitySource := []string{workspaceId}
	userIdentitySourceInvalidToken := []string{"eyJra.some.random.string"}
	datasetIdAsOrgIdQueryParam := map[string]string{"organization_id": datasetId, "someOtherParam": "someOtherValue"}
	orgIdAsDatasetIdQueryParam := map[string]string{"dataset_id": workspaceId, "someOtherParam": "someOtherValue"}
	manifestIdAsDatasetIdQueryParam := map[string]string{"dataset_id": manifestId, "someOtherParam": "someOtherValue"}
	datasetIdAsManifestIdQueryParam := map[string]string{"manifest_id": datasetId, "someOtherParam": "someOtherValue"}

	// error tests
	for scenario, params := range map[string]struct {
		idSource          []string
		queryParams       map[string]string
		expectedErrorText string
	}{
		"missing user token": {missingUserTokenIdentitySource, withWorkspaceId, "no suitable authorizer to process request"},
		"invalid user token": {userIdentitySourceInvalidToken, withoutIdQueryParams, "no suitable authorizer to process request"},
		"user supplies dataset id value to organization_id param when using a workspace authorizer endpoint": {datasetIdentitySource, datasetIdAsOrgIdQueryParam, "no suitable authorizer to process request"},
		"user supplies organization id value to dataset_id param when using a dataset authorizer endpoint":   {workspaceIdentitySource, orgIdAsDatasetIdQueryParam, "no suitable authorizer to process request"},
		"user supplies dataset id value to manifest_id when using a manifest authorizer endpoint":            {datasetIdentitySource, datasetIdAsManifestIdQueryParam, "no suitable authorizer to process request"},
		"user supplies a manifest id value to dataset_id when using a dataset authorizer endpoint":           {manifestIdentitySource, manifestIdAsDatasetIdQueryParam, "no suitable authorizer to process request"},
	} {
		t.Run(scenario, func(t *testing.T) {
			authorizer, err := authFactory.Build(params.idSource, params.queryParams)
			assert.Equal(t, nil, authorizer)
			assert.ErrorContains(t, err, params.expectedErrorText)
		})
	}
}
