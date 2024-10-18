package factory_test

import (
	"fmt"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/factory"
	"github.com/stretchr/testify/assert"
)

func TestFactory(t *testing.T) {
	authFactory := factory.NewCustomAuthorizerFactory()

	UserIdentitySource := []string{"UserAuthorizer", "Bearer eyJra.some.random.string"}
	authorizer, _ := authFactory.Build(UserIdentitySource)
	assert.Equal(t, fmt.Sprintf("%T", authorizer), "*authorizers.UserAuthorizer")

	DatasetIdentitySource := []string{"DatasetAuthorizer", "Bearer eyJra.some.random.string", "N:dataset:some-uuid"}
	authorizer, _ = authFactory.Build(DatasetIdentitySource)
	assert.Equal(t, fmt.Sprintf("%T", authorizer), "*authorizers.DatasetAuthorizer")

	ManifestIdentitySource := []string{"ManifestAuthorizer", "Bearer eyJra.some.random.string", "someManifestId"}
	authorizer, _ = authFactory.Build(ManifestIdentitySource)
	assert.Equal(t, fmt.Sprintf("%T", authorizer), "*authorizers.ManifestAuthorizer")

	WorkspaceIdentitySource := []string{"WorkspaceAuthorizer", "Bearer eyJra.some.random.string", "N:organization:some-uuid"}
	authorizer, _ = authFactory.Build(WorkspaceIdentitySource)
	assert.Equal(t, fmt.Sprintf("%T", authorizer), "*authorizers.WorkspaceAuthorizer")

}
