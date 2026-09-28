package manager_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test"
	pgModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Seed organizations.
const (
	org2NodeId = "N:organization:320813c5-3ea3-4c3b-aca5-9c6221e8d5f8"
	org3NodeId = "N:organization:4fb6fec6-9b2e-4885-91ff-7b3cf6579cd0"
	org4NodeId = "N:organization:8f60b0fd-55b7-4efa-b1b1-8204111117d3"
)

func addTeam(t *testing.T, db *sql.DB, orgId, userId int64) string {
	nodeId := fmt.Sprintf("N:team:%s", uuid.NewString())
	var teamId int64
	require.NoError(t, db.QueryRow(`INSERT INTO pennsieve.teams (name, node_id) VALUES ($1, $2) RETURNING id`,
		uuid.NewString(), nodeId).Scan(&teamId))
	t.Cleanup(func() {
		_, _ = db.Exec(`DELETE FROM pennsieve.team_user WHERE team_id = $1`, teamId)
		_, _ = db.Exec(`DELETE FROM pennsieve.organization_team WHERE team_id = $1`, teamId)
		_, _ = db.Exec(`DELETE FROM pennsieve.teams WHERE id = $1`, teamId)
	})
	_, err := db.Exec(`INSERT INTO pennsieve.organization_team (organization_id, team_id, permission_bit) VALUES ($1, $2, 2)`, orgId, teamId)
	require.NoError(t, err)
	if userId != 0 {
		_, err = db.Exec(`INSERT INTO pennsieve.team_user (team_id, user_id, permission_bit) VALUES ($1, $2, 2)`, teamId, userId)
		require.NoError(t, err)
	}
	return nodeId
}

func TestGetAccessScope(t *testing.T) {
	db, err := pgdb.ConnectENV()
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	require.NoError(t, db.Ping())

	user := test.NewUser(201, 2)
	test.AddUser(t, db, user, uuid.NewString())
	t.Cleanup(func() { test.DeleteUser(t, db, user.Id) })

	test.AddOrgUser(t, db, 2, user.Id, pgModels.Read)
	test.AddOrgUser(t, db, 3, user.Id, pgModels.Guest)
	test.AddOrgUser(t, db, 4, user.Id, pgModels.NoPermission)

	team2 := addTeam(t, db, 2, user.Id)
	team3 := addTeam(t, db, 3, user.Id)
	addTeam(t, db, 2, 0)       // a team the user isn't on
	addTeam(t, db, 4, user.Id) // a team in a workspace the user has no permission in

	ctx := context.Background()

	t.Run("all workspaces", func(t *testing.T) {
		scope, err := manager.GetAccessScope(ctx, db, user.Id, 0)
		require.NoError(t, err)
		assert.Equal(t, []string{org2NodeId, org3NodeId}, scope.WorkspaceNodeIds)
		assert.ElementsMatch(t, []string{team2, team3}, scope.TeamNodeIds)
	})

	t.Run("limited to an API token's workspace", func(t *testing.T) {
		scope, err := manager.GetAccessScope(ctx, db, user.Id, 3)
		require.NoError(t, err)
		assert.Equal(t, []string{org3NodeId}, scope.WorkspaceNodeIds)
		assert.Equal(t, []string{team3}, scope.TeamNodeIds)
	})

	t.Run("token workspace without permission", func(t *testing.T) {
		scope, err := manager.GetAccessScope(ctx, db, user.Id, 4)
		require.NoError(t, err)
		assert.Empty(t, scope.WorkspaceNodeIds)
		assert.Empty(t, scope.TeamNodeIds)
	})

	t.Run("unknown user", func(t *testing.T) {
		scope, err := manager.GetAccessScope(ctx, db, 999999, 0)
		require.NoError(t, err)
		assert.NotNil(t, scope.WorkspaceNodeIds, "empty lists, not null, in the Lambda request")
		assert.Empty(t, scope.WorkspaceNodeIds)
		assert.Empty(t, scope.TeamNodeIds)
	})
}
