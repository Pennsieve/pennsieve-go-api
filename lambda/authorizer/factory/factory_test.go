package factory_test

import (
	"fmt"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/factory"
	"github.com/stretchr/testify/assert"
)

func TestFactory(t *testing.T) {
	authFactory := factory.NewCustomAuthorizerFactory()

	withManifestId := map[string]string{"manifest_id": "someManifestId"}
	withDatasetId := map[string]string{"dataset_id": "someDatasetId"}
	withWorkspaceId := map[string]string{"dataset_id": "somehWorkspaceId"}
	withoutQueryParams := map[string]string{}

	UserIdentitySource := []string{"Bearer eyJra.some.random.string"}
	authorizer, _ := authFactory.Build(UserIdentitySource, withoutQueryParams)
	assert.Equal(t, "*authorizers.UserAuthorizer", fmt.Sprintf("%T", authorizer))

	DatasetIdentitySource := []string{"Bearer eyJra.some.random.string", "N:dataset:some-uuid"}
	authorizer, _ = authFactory.Build(DatasetIdentitySource, withDatasetId)
	assert.Equal(t, "*authorizers.DatasetAuthorizer", fmt.Sprintf("%T", authorizer))

	DatasetIdentitySourceFlippedOrder := []string{"N:dataset:some-uuid", "Bearer eyJra.some.random.string"}
	authorizer, _ = authFactory.Build(DatasetIdentitySourceFlippedOrder, withDatasetId)
	assert.Equal(t, "*authorizers.DatasetAuthorizer", fmt.Sprintf("%T", authorizer))

	ManifestIdentitySource := []string{"Bearer eyJra.some.random.string", "someManifestId"}
	authorizer, _ = authFactory.Build(ManifestIdentitySource, withManifestId)
	assert.Equal(t, "*authorizers.ManifestAuthorizer", fmt.Sprintf("%T", authorizer))

	ManifestIdentitySourceFlippedOrder := []string{"someManifestId", "Bearer eyJra.some.random.string"}
	authorizer, _ = authFactory.Build(ManifestIdentitySourceFlippedOrder, withManifestId)
	assert.Equal(t, "*authorizers.ManifestAuthorizer", fmt.Sprintf("%T", authorizer))

	WorkspaceIdentitySource := []string{"Bearer eyJra.some.random.string", "N:organization:some-uuid"}
	authorizer, _ = authFactory.Build(WorkspaceIdentitySource, withWorkspaceId)
	assert.Equal(t, "*authorizers.WorkspaceAuthorizer", fmt.Sprintf("%T", authorizer))

	WorkspaceIdentitySourceFlippedOrder := []string{"N:organization:some-uuid", "Bearer eyJra.some.random.string"}
	authorizer, _ = authFactory.Build(WorkspaceIdentitySourceFlippedOrder, withWorkspaceId)
	assert.Equal(t, "*authorizers.WorkspaceAuthorizer", fmt.Sprintf("%T", authorizer))

	InvalidIdentitySource := []string{"N:organization:some-uuid"}
	authorizer, err := authFactory.Build(InvalidIdentitySource, withWorkspaceId)
	assert.Equal(t, nil, authorizer)
	assert.Equal(t, err.Error(), "no suitable authorizer to process request")

	UserIdentitySourceInvalidToken := []string{"eyJra.some.random.string"}
	authorizer, err = authFactory.Build(UserIdentitySourceInvalidToken, withoutQueryParams)
	assert.Equal(t, nil, authorizer)
	assert.Equal(t, err.Error(), "no suitable authorizer to process request")

}
