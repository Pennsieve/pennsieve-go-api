package mappers_test

import (
	"fmt"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/mappers"
	"github.com/stretchr/testify/assert"
)

func TestMappers(t *testing.T) {
	authFactory := mappers.NewCustomAuthorizerFactory(nil, nil, nil)

	UserIdentitySource := []string{"Bearer eyJra.some.random.string"}
	authorizer, _ := mappers.IdentitySourceToAuthorizer(UserIdentitySource, authFactory)
	assert.Equal(t, fmt.Sprintf("%T", authorizer), "*authorizers.UserAuthorizer")

	DatasetIdentitySource := []string{"Bearer eyJra.some.random.string", "N:dataset:some-uuid"}
	authorizer, _ = mappers.IdentitySourceToAuthorizer(DatasetIdentitySource, authFactory)
	assert.Equal(t, fmt.Sprintf("%T", authorizer), "*authorizers.DatasetAuthorizer")

	ManifestIdentitySource := []string{"Bearer eyJra.some.random.string", "N:manifest:some-uuid"}
	authorizer, _ = mappers.IdentitySourceToAuthorizer(ManifestIdentitySource, authFactory)
	assert.Equal(t, fmt.Sprintf("%T", authorizer), "*authorizers.ManifestAuthorizer")

	WorkspaceIdentitySource := []string{"Bearer eyJra.some.random.string", "N:organization:some-uuid"}
	authorizer, _ = mappers.IdentitySourceToAuthorizer(WorkspaceIdentitySource, authFactory)
	assert.Equal(t, fmt.Sprintf("%T", authorizer), "*authorizers.WorkspaceAuthorizer")

}
