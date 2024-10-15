package mappers_test

import (
	"context"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/mappers"
	"github.com/stretchr/testify/assert"
)

func TestMappers(t *testing.T) {
	UserIdentitySource := []string{"Bearer eyJra.some.random.string"}
	authorizer, _ := mappers.IdentitySourceToAuthorizer(UserIdentitySource, nil, nil, nil)
	result, _ := authorizer.GenerateClaims(context.Background())
	assert.Equal(t, len(result), 1)

	// DatasetIdentitySource := []string{"Bearer eyJra.some.random.string", "N:dataset:some-uuid"}
	// authorizer = mappers.IdentitySourceToAuthorizer(DatasetIdentitySource)
	// assert.Equal(t, len(authorizer.GenerateClaims()), 2)
	// ManifestIdentitySource := []string{"Bearer eyJra.some.random.string", "N:manifest:some-uuid"}
	// authorizer = mappers.IdentitySourceToAuthorizer(ManifestIdentitySource)
	// assert.Equal(t, len(authorizer.GenerateClaims()), 2)

	// WorkspaceIdentitySource := []string{"Bearer eyJra.some.random.string", "N:organization:some-uuid"}
	// authorizer = mappers.IdentitySourceToAuthorizer(WorkspaceIdentitySource)
	// assert.Equal(t, len(authorizer.GenerateClaims()), 3)

}
