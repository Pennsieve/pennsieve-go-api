package authorizers_test

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	pgModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

var seedOrgIdToNodeId = map[int64]string{
	2: "N:organization:320813c5-3ea3-4c3b-aca5-9c6221e8d5f8",
	3: "N:organization:4fb6fec6-9b2e-4885-91ff-7b3cf6579cd0",
	4: "N:organization:8f60b0fd-55b7-4efa-b1b1-8204111117d3",
}

type userWithCognito struct {
	user            *pgModels.User
	cognitoUsername string
}

func (u userWithCognito) delete(t require.TestingT, db *sql.DB) {
	test.DeleteUser(t, db, u.user.Id)
}

var testUser = userWithCognito{
	// deliberately use a preferredOrgId that doest not exist in the seed DB since this should be ignored
	// by the workspace authorizer
	user:            test.NewUser(101, 1001),
	cognitoUsername: uuid.NewString(),
}

func TestWorkspaceAuthorizer(t *testing.T) {
	pgDB, err := pgdb.ConnectENV()
	require.NoError(t, err)
	t.Cleanup(func() {
		if err := pgDB.Close(); err != nil {
			t.Log("error closing test Postgres DB:", err)
		}
	})
	require.NoError(t, pgDB.Ping())
	for scenario, testFunc := range map[string]func(*testing.T, *sql.DB){
		"user not in workspace":                        testUserNotInWorkspace,
		"api token does not match requested workspace": testAPIKeyNotInRequestedWorkspace,
		"user with delete perm in workspace":           testUserWithDeleteInWorkspace,
		"api token with delete perm in workspace":      testAPIKeyWithDeleteInWorkspace,
		"user with no permissions perm in workspace":   testUserWithNoPermissionInWorkspace,
	} {
		t.Run(scenario, func(t *testing.T) {
			// subtests may add rows to organization_user and tokens tables, so we add
			// the user before each test so that we can delete it after each test which
			// will cascade that delete to those other tables
			test.AddUser(t, pgDB, testUser.user, testUser.cognitoUsername)
			t.Cleanup(func() {
				testUser.delete(t, pgDB)
			})
			testFunc(t, pgDB)
		})
	}
}

func testUserNotInWorkspace(t *testing.T, pgDB *sql.DB) {
	pgQueries := pgdb.New(pgDB)
	token := test.NewJWTBuilder().WithUsername(testUser.cognitoUsername).Build(t)
	claimsManager := manager.NewClaimsManager(pgQueries, nil, token.Token, uuid.NewString(), uuid.NewString())

	workspaceAuthorizer := authorizers.NewWorkspaceAuthorizer(seedOrgIdToNodeId[2])

	_, err := workspaceAuthorizer.GenerateClaims(context.Background(), claimsManager, "")
	assert.ErrorContains(t, err, "organization user was not found")
}

func testAPIKeyNotInRequestedWorkspace(t *testing.T, pgDB *sql.DB) {
	pgQueries := pgdb.New(pgDB)

	// API token that is for seed workspace 2
	tokenWorkspaceId := int64(2)
	tokenWorkspaceNodeId := seedOrgIdToNodeId[tokenWorkspaceId]
	token := test.NewJWTBuilder().WithUsername(testUser.cognitoUsername).WithWorkspace(tokenWorkspaceId, tokenWorkspaceNodeId).Build(t)

	test.AddOrgUser(t, pgDB, tokenWorkspaceId, testUser.user.Id, pgModels.Owner)
	test.AddAPIToken(t, pgDB, tokenWorkspaceId, testUser.user.Id, testUser.cognitoUsername, token.ClientId)
	claimsManager := manager.NewClaimsManager(pgQueries, nil, token.Token, token.ClientId, uuid.NewString())

	// workspace authorizer for seed workspace 3
	authorizerWorkspaceNodeId := seedOrgIdToNodeId[3]
	workspaceAuthorizer := authorizers.NewWorkspaceAuthorizer(authorizerWorkspaceNodeId)

	_, err := workspaceAuthorizer.GenerateClaims(context.Background(), claimsManager, "")
	assert.ErrorContains(t, err,
		fmt.Sprintf("provided workspace id %s does not match API token workspace id %s",
			authorizerWorkspaceNodeId, tokenWorkspaceNodeId))
}

func testUserWithDeleteInWorkspace(t *testing.T, pgDB *sql.DB) {
	orgId := int64(3)
	orgNodeId := seedOrgIdToNodeId[orgId]
	test.AddOrgUser(t, pgDB, orgId, testUser.user.Id, pgModels.Delete)

	pgQueries := pgdb.New(pgDB)
	token := test.NewJWTBuilder().WithUsername(testUser.cognitoUsername).Build(t)

	claimsManager := manager.NewClaimsManager(pgQueries, nil, token.Token, uuid.NewString(), uuid.NewString())

	workspaceAuthorizer := authorizers.NewWorkspaceAuthorizer(orgNodeId)
	claims, err := workspaceAuthorizer.GenerateClaims(context.Background(), claimsManager, "")
	assert.NoError(t, err)

	assert.Len(t, claims, 3)
	assert.Contains(t, claims, "org_claim")

	var orgClaim *organization.Claim
	require.IsType(t, orgClaim, claims["org_claim"])
	orgClaim = claims["org_claim"].(*organization.Claim)
	assert.Equal(t, orgNodeId, orgClaim.NodeId)
	assert.Equal(t, orgId, orgClaim.IntId)
	assert.Equal(t, pgModels.Delete, orgClaim.Role)

}

func testUserWithNoPermissionInWorkspace(t *testing.T, pgDB *sql.DB) {
	orgId := int64(3)
	orgNodeId := seedOrgIdToNodeId[orgId]
	test.AddOrgUser(t, pgDB, orgId, testUser.user.Id, pgModels.NoPermission)

	pgQueries := pgdb.New(pgDB)
	token := test.NewJWTBuilder().WithUsername(testUser.cognitoUsername).Build(t)

	claimsManager := manager.NewClaimsManager(pgQueries, nil, token.Token, uuid.NewString(), uuid.NewString())

	workspaceAuthorizer := authorizers.NewWorkspaceAuthorizer(orgNodeId)
	claims, err := workspaceAuthorizer.GenerateClaims(context.Background(), claimsManager, "")
	// TODO :should this return an error if the user has no permissions in the DB?
	assert.NoError(t, err)

	assert.Len(t, claims, 3)
	assert.Contains(t, claims, "org_claim")

	var orgClaim *organization.Claim
	require.IsType(t, orgClaim, claims["org_claim"])
	orgClaim = claims["org_claim"].(*organization.Claim)
	assert.Equal(t, orgNodeId, orgClaim.NodeId)
	assert.Equal(t, orgId, orgClaim.IntId)
	assert.Equal(t, pgModels.NoPermission, orgClaim.Role)

}

func testAPIKeyWithDeleteInWorkspace(t *testing.T, pgDB *sql.DB) {
	orgId := int64(3)
	orgNodeId := seedOrgIdToNodeId[orgId]
	token := test.NewJWTBuilder().
		WithUsername(testUser.cognitoUsername).
		WithWorkspace(orgId, orgNodeId).
		Build(t)
	test.AddOrgUser(t, pgDB, orgId, testUser.user.Id, pgModels.Delete)
	test.AddAPIToken(t, pgDB, orgId, testUser.user.Id, testUser.cognitoUsername, token.ClientId)

	pgQueries := pgdb.New(pgDB)

	claimsManager := manager.NewClaimsManager(pgQueries, nil, token.Token, token.ClientId, uuid.NewString())

	workspaceAuthorizer := authorizers.NewWorkspaceAuthorizer(orgNodeId)
	claims, err := workspaceAuthorizer.GenerateClaims(context.Background(), claimsManager, "")
	assert.NoError(t, err)

	assert.Len(t, claims, 3)
	assert.Contains(t, claims, "org_claim")

	var orgClaim *organization.Claim
	require.IsType(t, orgClaim, claims["org_claim"])
	orgClaim = claims["org_claim"].(*organization.Claim)
	assert.Equal(t, orgNodeId, orgClaim.NodeId)
	assert.Equal(t, orgId, orgClaim.IntId)
	assert.Equal(t, pgModels.Delete, orgClaim.Role)

}
