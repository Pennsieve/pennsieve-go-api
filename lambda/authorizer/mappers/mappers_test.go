package mappers_test

import (
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/mappers"
	"github.com/stretchr/testify/assert"
)

func TestMappers(t *testing.T) {
	DatasetIdentitySource := []string{"Bearer eyJra.some.random.string", "N:dataset:some-uuid"}
	authorizer := mappers.IdentitySourceToAuthorizer(DatasetIdentitySource)
	assert.Equal(t, len(authorizer.GenerateClaims()), 2)

	WorkspaceIdentitySource := []string{"Bearer eyJra.some.random.string", "N:organization:some-uuid"}
	authorizer = mappers.IdentitySourceToAuthorizer(WorkspaceIdentitySource)
	assert.Equal(t, len(authorizer.GenerateClaims()), 3)

	UserIdentitySource := []string{"Bearer eyJra.some.random.string"}
	authorizer = mappers.IdentitySourceToAuthorizer(UserIdentitySource)
	assert.Equal(t, len(authorizer.GenerateClaims()), 1)
}
