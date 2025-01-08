package manager_test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test/mocks"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dydb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/teamUser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type managerParams struct {
	mockPennsievePg   *mocks.MockPennsievePgAPI
	mockPennsieveDy   *mocks.MockPennsieveDyAPI
	testJWT           test.JWT
	tokenClientId     string
	manifestTableName string
}

func (p *managerParams) buildManager() manager.IdentityManager {
	return manager.NewClaimsManager(p.mockPennsievePg, p.mockPennsieveDy, p.testJWT.Token, p.tokenClientId, p.manifestTableName)
}

func (p *managerParams) withUserQueryMocked(t require.TestingT, currentUser *pgdb.User) *managerParams {
	if p.testJWT.Workspace == nil && p.tokenClientId != p.testJWT.ClientId {
		// If the jwt does not contain a workspace and did not come from the token pool, then we
		// expect the manager.ClaimsManager to call the pgdb method that queries only the user table
		p.mockPennsievePg.OnGetByCognitoId(p.testJWT.Username).Return(currentUser, nil)
	} else if p.testJWT.Workspace != nil && p.tokenClientId == p.testJWT.ClientId {
		// If the jwt contains a workspace and came from the token pool, then we
		// expect the manager.ClaimsManager to call the pgdb method that queries a join of the users and token tables.
		p.mockPennsievePg.OnGetUserByCognitoId(p.testJWT.Username).Return(currentUser, nil)
	} else {
		require.FailNow(t, "inconsistent managerParams", "testJWT should be non-nil if and only if testJWT clientId is the tokenClientId")
	}
	return p
}

func (p *managerParams) getExpectedOrgId(user *pgdb.User) int64 {
	if p.testJWT.Workspace == nil {
		return user.PreferredOrg
	}
	return p.testJWT.Workspace.Id
}

func (p *managerParams) getExpectedOrgNodeId() string {
	if p.testJWT.Workspace == nil {
		return fmt.Sprintf("N:organization:%s", uuid.NewString())
	}
	return p.testJWT.Workspace.NodeId
}

func (p *managerParams) assertMockExpectations(t *testing.T) {
	p.mockPennsievePg.AssertExpectations(t)
	p.mockPennsieveDy.AssertExpectations(t)
}

func newNoWorkspaceTokenManagerParams(t require.TestingT) *managerParams {
	// A JWT token with no workspace and a clientId
	// different from the tokenClientId (initialized below)
	testJWT := test.NewJWTBuilder().Build(t)

	return &managerParams{
		mockPennsievePg: mocks.NewMockPennsievePgAPI(),
		mockPennsieveDy: mocks.NewMockPennsieveDyAPI(),
		testJWT:         testJWT,
		// tokenClientID will be different from the random clientId in testJWT
		tokenClientId:     uuid.NewString(),
		manifestTableName: uuid.NewString(),
	}
}

func newWorkspaceTokenManagerParams(t require.TestingT, tokenWorkspace manager.TokenWorkspace) *managerParams {
	// A JWT token with a workspace and a clientId
	// that will match the tokenClientId (initialized below)
	testJWT := test.NewJWTBuilder().
		WithWorkspace(tokenWorkspace.Id, tokenWorkspace.NodeId).
		Build(t)

	return &managerParams{
		mockPennsievePg: mocks.NewMockPennsievePgAPI(),
		mockPennsieveDy: mocks.NewMockPennsieveDyAPI(),
		testJWT:         testJWT,
		// tokenClientID matches the clientId in token
		tokenClientId:     testJWT.ClientId,
		manifestTableName: uuid.NewString(),
	}
}

func TestClaimsManager(t *testing.T) {

	for scenario, tstFunc := range map[string]func(t *testing.T, params *managerParams){
		"GetCurrentUser":      testGetCurrentUser,
		"GetUserClaim":        testGetUserClaim,
		"GetTokenWorkspace":   testGetTokenWorkspace,
		"GetActiveOrg":        testGetActiveOrg,
		"GetDatasetClaim":     testGetDatasetClaim,
		"GetDatasetId":        testGetDatasetId,
		"GetOrgClaim":         testGetOrgClaim,
		"GetOrgClaimByNodeId": testGetOrgClaimByNodeId,
		"GetTeamClaims":       testGetTeamClaims,
	} {
		t.Run(scenario, func(t *testing.T) {

			t.Run("token without workspace", func(t *testing.T) {
				noWorkspaceParams := newNoWorkspaceTokenManagerParams(t)
				tstFunc(t, noWorkspaceParams)
				noWorkspaceParams.assertMockExpectations(t)
			})

			t.Run("token with workspace", func(t *testing.T) {
				tokenWorkspace := manager.TokenWorkspace{
					Id:     5001,
					NodeId: fmt.Sprintf("N:organization:%s", uuid.NewString()),
				}
				withWorkspaceParams := newWorkspaceTokenManagerParams(t, tokenWorkspace)
				tstFunc(t, withWorkspaceParams)
				withWorkspaceParams.assertMockExpectations(t)
			})

		})
	}

}

func testGetCurrentUser(t *testing.T, params *managerParams) {
	expectedUser := test.NewUser(101, 2001)
	claimsManager := params.withUserQueryMocked(t, expectedUser).buildManager()

	ctx := context.Background()
	user, err := claimsManager.GetCurrentUser(ctx)
	require.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func testGetUserClaim(t *testing.T, params *managerParams) {
	claimsManager := params.buildManager()

	expectedUser := test.NewUser(101, 2001)
	ctx := context.Background()

	userClaim := claimsManager.GetUserClaim(ctx, expectedUser)
	assert.Equal(t, expectedUser.Id, userClaim.Id)
	assert.Equal(t, expectedUser.NodeId, userClaim.NodeId)
	assert.Equal(t, expectedUser.IsSuperAdmin, userClaim.IsSuperAdmin)

}

func testGetTokenWorkspace(t *testing.T, params *managerParams) {
	claimsManager := params.buildManager()

	tokenWorkspace, hasTokenWorkspace := claimsManager.GetTokenWorkspace()
	if params.testJWT.Workspace == nil {
		assert.False(t, hasTokenWorkspace)
	} else {
		assert.True(t, hasTokenWorkspace)
		assert.Equal(t, *params.testJWT.Workspace, tokenWorkspace)
	}
}

func testGetActiveOrg(t *testing.T, params *managerParams) {
	claimsManager := params.buildManager()

	expectedUser := test.NewUser(101, 2001)
	expectedOrgId := params.getExpectedOrgId(expectedUser)

	ctx := context.Background()
	orgId := claimsManager.GetActiveOrg(ctx, expectedUser)
	assert.Equal(t, expectedOrgId, orgId)
	if params.testJWT.Workspace != nil {
		assert.NotEqual(t, expectedUser.PreferredOrg, orgId)
	}
}

func testGetDatasetClaim(t *testing.T, params *managerParams) {
	claimsManager := params.buildManager()

	// Set up Mock
	expectedUser := test.NewUser(101, 2001)
	datasetId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	expectedOrgId := params.getExpectedOrgId(expectedUser)
	expectedDatasetClaim := &dataset.Claim{
		Role:   role.Viewer,
		NodeId: datasetId,
		IntId:  555,
	}
	params.mockPennsievePg.OnGetDatasetClaim(expectedUser, datasetId, expectedOrgId).Return(expectedDatasetClaim, nil)

	ctx := context.Background()
	claim, err := claimsManager.GetDatasetClaim(ctx, expectedUser, datasetId, expectedOrgId)
	require.NoError(t, err)

	assert.Equal(t, expectedDatasetClaim, claim)
}

func testGetDatasetId(t *testing.T, params *managerParams) {
	claimsManager := params.buildManager()

	// set up mock
	expectedManifestId := uuid.NewString()
	expectedManifest := &dydb.ManifestTable{
		ManifestId:    expectedManifestId,
		DatasetId:     555,
		DatasetNodeId: fmt.Sprintf("N:dataset:%s", uuid.NewString()),
	}
	params.mockPennsieveDy.OnGetManifestById(params.manifestTableName, expectedManifestId).Return(expectedManifest, nil)

	ctx := context.Background()
	datasetId, err := claimsManager.GetDatasetID(ctx, expectedManifestId)
	require.NoError(t, err)
	assert.Equal(t, expectedManifest.DatasetNodeId, datasetId)
}

func testGetOrgClaim(t *testing.T, params *managerParams) {
	claimsManager := params.buildManager()

	expectedUser := test.NewUser(101, 2001)
	expectedOrgId := params.getExpectedOrgId(expectedUser)
	expectedOrgNodeId := params.getExpectedOrgNodeId()
	expectedClaim := &organization.Claim{
		Role:            pgdb.Owner,
		IntId:           expectedOrgId,
		NodeId:          expectedOrgNodeId,
		EnabledFeatures: []pgdb.FeatureFlags{{OrganizationId: expectedOrgId, Feature: "test-feature", Enabled: true}},
	}
	params.mockPennsievePg.OnGetOrganizationClaim(expectedUser.Id, expectedOrgId).Return(expectedClaim, nil)

	ctx := context.Background()
	claim, err := claimsManager.GetOrgClaim(ctx, expectedUser.Id, expectedOrgId)
	require.NoError(t, err)
	assert.Equal(t, expectedClaim, claim)
}

func testGetOrgClaimByNodeId(t *testing.T, params *managerParams) {
	claimsManager := params.buildManager()

	expectedUser := test.NewUser(101, 2001)
	expectedOrgId := params.getExpectedOrgId(expectedUser)
	expectedOrgNodeId := params.getExpectedOrgNodeId()
	expectedClaim := &organization.Claim{
		Role:            pgdb.Owner,
		IntId:           expectedOrgId,
		NodeId:          expectedOrgNodeId,
		EnabledFeatures: []pgdb.FeatureFlags{{OrganizationId: expectedOrgId, Feature: "test-feature", Enabled: true}},
	}
	params.mockPennsievePg.OnGetOrganizationClaimByNodeId(expectedUser.Id, expectedOrgNodeId).Return(expectedClaim, nil)

	ctx := context.Background()
	claim, err := claimsManager.GetOrgClaimByNodeId(ctx, expectedUser.Id, expectedOrgNodeId)
	require.NoError(t, err)
	assert.Equal(t, expectedClaim, claim)
}

func testGetTeamClaims(t *testing.T, params *managerParams) {
	claimsManager := params.buildManager()

	expectedUser := test.NewUser(101, 2001)
	expectedClaims := []teamUser.Claim{
		{
			IntId:      10,
			Name:       "team 1",
			NodeId:     fmt.Sprintf("N:team:%s", uuid.NewString()),
			Permission: pgdb.Guest,
			TeamType:   "type 1",
		},
		{
			IntId:      21,
			Name:       "team 2",
			NodeId:     fmt.Sprintf("N:team:%s", uuid.NewString()),
			Permission: pgdb.Administer,
			TeamType:   "type 2",
		},
	}
	params.mockPennsievePg.OnGetTeamClaims(expectedUser.Id).Return(expectedClaims, nil)

	ctx := context.Background()
	claims, err := claimsManager.GetTeamClaims(ctx, expectedUser.Id)
	require.NoError(t, err)
	assert.Equal(t, expectedClaims, claims)
}
