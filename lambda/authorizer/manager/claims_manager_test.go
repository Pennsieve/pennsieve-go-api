package manager

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestClaimsManager_GetUserTokenWorkspace(t *testing.T) {
	orgId := int64(13)
	orgNodeId := fmt.Sprintf("N:organization:%s", uuid.NewString())
	jwtBuilder := jwt.NewBuilder().
		Claim("username", uuid.NewString()).
		Claim("client_id", uuid.NewString())

	tokenWithoutOrg, err := jwtBuilder.Build()
	require.NoError(t, err)

	t.Run("token without workspace", func(t *testing.T) {
		manager := NewClaimsManager(nil, nil, tokenWithoutOrg, uuid.NewString())
		_, hasWorkspace := manager.GetUserTokenWorkspace()
		assert.False(t, hasWorkspace)
	})

	tokenWithOrg, err := jwtBuilder.
		Claim("custom:organization_id", orgId).
		Claim("custom:organization_node_id", orgNodeId).
		Build()
	require.NoError(t, err)

	t.Run("token with workspace", func(t *testing.T) {
		manager := NewClaimsManager(nil, nil, tokenWithOrg, uuid.NewString())
		tokenWorkspace, hasWorkspace := manager.GetUserTokenWorkspace()
		assert.True(t, hasWorkspace)
		assert.Equal(t, orgId, tokenWorkspace.Id)
		assert.Equal(t, orgNodeId, tokenWorkspace.NodeId)
	})
}
