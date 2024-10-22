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
	assert.Equal(t, fmt.Sprintf("%T", authorizer), "*authorizers.UserAuthorizer")

	DatasetIdentitySource := []string{"Bearer eyJra.some.random.string", "N:dataset:some-uuid"}
	authorizer, _ = authFactory.Build(DatasetIdentitySource, withDatasetId)
	assert.Equal(t, fmt.Sprintf("%T", authorizer), "*authorizers.DatasetAuthorizer")

	ManifestIdentitySource := []string{"Bearer eyJra.some.random.string", "someManifestId"}
	authorizer, _ = authFactory.Build(ManifestIdentitySource, withManifestId)
	assert.Equal(t, fmt.Sprintf("%T", authorizer), "*authorizers.ManifestAuthorizer")

	WorkspaceIdentitySource := []string{"Bearer eyJra.some.random.string", "N:organization:some-uuid"}
	authorizer, _ = authFactory.Build(WorkspaceIdentitySource, withWorkspaceId)
	assert.Equal(t, fmt.Sprintf("%T", authorizer), "*authorizers.WorkspaceAuthorizer")

}
