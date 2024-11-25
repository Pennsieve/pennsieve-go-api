package manager_test

import (
	"context"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestClaimsManager_GetTokenWorkspace(t *testing.T) {
	t.Run("token without workspace", func(t *testing.T) {
		tokenWithoutOrg := test.NewJWT(t)
		claimsManager := manager.NewClaimsManager(nil, nil, tokenWithoutOrg.Token, uuid.NewString(), uuid.NewString())
		_, hasWorkspace := claimsManager.GetTokenWorkspace()
		assert.False(t, hasWorkspace)
	})

	t.Run("token with workspace", func(t *testing.T) {
		tokenWithOrg := test.NewJWTWithWorkspace(t)
		claimsManager := manager.NewClaimsManager(nil, nil, tokenWithOrg.Token, uuid.NewString(), uuid.NewString())
		tokenWorkspace, hasWorkspace := claimsManager.GetTokenWorkspace()
		assert.True(t, hasWorkspace)
		assert.Equal(t, tokenWithOrg.Workspace.Id, tokenWorkspace.Id)
		assert.Equal(t, tokenWithOrg.Workspace.NodeId, tokenWorkspace.NodeId)
	})
}

func TestClaimsManager_GetActiveOrg(t *testing.T) {
	user := test.NewUser(101, 1001) // Greater org id than will be returned by NewTestJWTWithWorkspace

	t.Run("token without workspace", func(t *testing.T) {
		tokenWithoutOrg := test.NewJWT(t)
		claimsManager := manager.NewClaimsManager(nil, nil, tokenWithoutOrg.Token, uuid.NewString(), uuid.NewString())
		orgId := claimsManager.GetActiveOrg(context.Background(), user)
		// GetActiveOrg returns the user's preferred org since the token does not contain a workspace
		assert.Equal(t, user.PreferredOrg, orgId)
	})

	t.Run("token with workspace", func(t *testing.T) {
		tokenWithOrg := test.NewJWTWithWorkspace(t)
		claimsManager := manager.NewClaimsManager(nil, nil, tokenWithOrg.Token, uuid.NewString(), uuid.NewString())
		orgId := claimsManager.GetActiveOrg(context.Background(), user)
		// GetActiveOrg returns the workspace from the JWT token and ignores the user's preferred org
		assert.Equal(t, tokenWithOrg.Workspace.Id, orgId)
	})
}
