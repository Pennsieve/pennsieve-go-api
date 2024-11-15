package mappers_test

import (
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/mappers"
	"github.com/stretchr/testify/assert"
)

func TestIdentitySourceMapper(t *testing.T) {
	UserIdentitySource := []string{"Bearer eyJra.some.random.string"}
	DatasetIdentitySource := []string{"Bearer eyJra.some.random.string", "N:dataset:some-uuid"}
	DatasetIdentitySourceFlippedOrder := []string{"N:dataset:some-uuid", "Bearer eyJra.some.random.string"}
	ManifestIdentitySource := []string{"Bearer eyJra.some.random.string", "someManifestId"}
	ManifestIdentitySourceFlippedOrder := []string{"someManifestId", "Bearer eyJra.some.random.string"}
	WorkspaceIdentitySource := []string{"Bearer eyJra.some.random.string", "N:organization:some-uuid"}
	WorkspaceIdentitySourceFlippedOrder := []string{"N:organization:some-uuid", "Bearer eyJra.some.random.string"}
	UserIdentitySourceInvalidToken := []string{"eyJra.some.random.string"}

	identitySourceMapper := mappers.NewIdentitySourceMapper(UserIdentitySource, false)
	auxiliaryIdentitySource := identitySourceMapper.Create()
	assert.Equal(t, "Bearer eyJra.some.random.string", auxiliaryIdentitySource["token"])
	assert.Equal(t, 1, len(auxiliaryIdentitySource))

	identitySourceMapper = mappers.NewIdentitySourceMapper(DatasetIdentitySource, false)
	auxiliaryIdentitySource = identitySourceMapper.Create()
	assert.Equal(t, "Bearer eyJra.some.random.string", auxiliaryIdentitySource["token"])
	assert.Equal(t, "N:dataset:some-uuid", auxiliaryIdentitySource["dataset_id"])
	assert.Equal(t, 2, len(auxiliaryIdentitySource))

	identitySourceMapper = mappers.NewIdentitySourceMapper(DatasetIdentitySourceFlippedOrder, false)
	auxiliaryIdentitySource = identitySourceMapper.Create()
	assert.Equal(t, "Bearer eyJra.some.random.string", auxiliaryIdentitySource["token"])
	assert.Equal(t, "N:dataset:some-uuid", auxiliaryIdentitySource["dataset_id"])
	assert.Equal(t, 2, len(auxiliaryIdentitySource))

	identitySourceMapper = mappers.NewIdentitySourceMapper(ManifestIdentitySource, true)
	auxiliaryIdentitySource = identitySourceMapper.Create()
	assert.Equal(t, "Bearer eyJra.some.random.string", auxiliaryIdentitySource["token"])
	assert.Equal(t, "someManifestId", auxiliaryIdentitySource["manifest_id"])
	assert.Equal(t, 2, len(auxiliaryIdentitySource))

	identitySourceMapper = mappers.NewIdentitySourceMapper(ManifestIdentitySourceFlippedOrder, true)
	auxiliaryIdentitySource = identitySourceMapper.Create()
	assert.Equal(t, "Bearer eyJra.some.random.string", auxiliaryIdentitySource["token"])
	assert.Equal(t, "someManifestId", auxiliaryIdentitySource["manifest_id"])
	assert.Equal(t, 2, len(auxiliaryIdentitySource))

	identitySourceMapper = mappers.NewIdentitySourceMapper(WorkspaceIdentitySource, false)
	auxiliaryIdentitySource = identitySourceMapper.Create()
	assert.Equal(t, "Bearer eyJra.some.random.string", auxiliaryIdentitySource["token"])
	assert.Equal(t, "N:organization:some-uuid", auxiliaryIdentitySource["workspace_id"])
	assert.Equal(t, 2, len(auxiliaryIdentitySource))

	identitySourceMapper = mappers.NewIdentitySourceMapper(WorkspaceIdentitySourceFlippedOrder, false)
	auxiliaryIdentitySource = identitySourceMapper.Create()
	assert.Equal(t, "Bearer eyJra.some.random.string", auxiliaryIdentitySource["token"])
	assert.Equal(t, "N:organization:some-uuid", auxiliaryIdentitySource["workspace_id"])
	assert.Equal(t, 2, len(auxiliaryIdentitySource))

	identitySourceMapper = mappers.NewIdentitySourceMapper(UserIdentitySourceInvalidToken, false)
	auxiliaryIdentitySource = identitySourceMapper.Create()
	assert.Equal(t, 0, len(auxiliaryIdentitySource))
}
