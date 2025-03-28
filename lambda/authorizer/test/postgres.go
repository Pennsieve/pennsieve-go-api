package test

import (
	"database/sql"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/stretchr/testify/require"
)

// AddUser inserts a user into the seed test database. The given user must have id > 3 since the seed database
// already has users 1, 2, and 3. And the users id sequence is not updated in the seed, so if you try and insert a user
// without an id it fails with a users.id uniqueness constraint.
func AddUser(t require.TestingT, db *sql.DB, user *pgdb.User, cognitoId string) {
	require.True(t, user.Id > 3, "test user id should be > 3 to avoid conflict with existing seed users")
	query := `INSERT INTO "pennsieve"."users" 
    						(id, email, first_name, last_name, is_super_admin, preferred_org_id, node_id, cognito_id) 
							VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`
	_, err := db.Exec(query, user.Id, user.Email, user.FirstName, user.LastName, user.IsSuperAdmin, user.PreferredOrg, user.NodeId, cognitoId)
	require.NoError(t, err, "error inserting test user")
}

func DeleteUser(t require.TestingT, db *sql.DB, userId int64) {
	query := `DELETE FROM "pennsieve"."users" WHERE id = $1`
	result, err := db.Exec(query, userId)
	require.NoError(t, err, "error deleting test user %d", userId)
	affectedCount, err := result.RowsAffected()
	require.NoError(t, err, "error checking affected row count when deleting test user %d", userId)
	require.Equal(t, int64(1), affectedCount, "expected to delete exactly one test user with id %d, actual deleted rows: %d", userId, affectedCount)
}

// AddOrgUser adds the given user to the given workspace with the given permissions
// If the given user is deleted, the row in organization_user will automatically be deleted because of
// a delete cascade in the seed DB's DDL.
func AddOrgUser(t require.TestingT, db *sql.DB, orgId, userId int64, orgPermissions pgdb.DbPermission) {
	query := `INSERT INTO 
    		  "pennsieve"."organization_user" (organization_id, user_id, permission_bit)
    		  VALUES ($1, $2, $3)`
	_, err := db.Exec(query, orgId, userId, orgPermissions)
	require.NoError(t, err, "error inserting (organization, user) (%d, %d)", orgId, userId)

}

// AddAPIToken adds an API token to the given user in the given organization.
// If the given user is deleted, the row in tokens will automatically be deleted because of
// a delete cascade in the seed DB's DDL.
func AddAPIToken(t require.TestingT, db *sql.DB, orgId, userId int64, apiKey string, clientId string) {
	query := `INSERT INTO 
    		  "pennsieve"."tokens" (name, token, organization_id, user_id, cognito_id)
    		  VALUES ($1, $2, $3, $4, $5)`
	_, err := db.Exec(query, uuid.NewString(), apiKey, orgId, userId, clientId)
	require.NoError(t, err, "error inserting (org, user, token, clientId) into tokens (%d, %d, %s, %s)",
		orgId, userId, apiKey, clientId)
}
