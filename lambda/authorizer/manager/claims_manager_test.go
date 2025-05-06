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

func TestClaimsManager(t *testing.T) {

	for scenario, tstFunc := range map[string]func(t *testing.T, params *mocks.ClaimsManagerParams){
		"GetCurrentUser":      testGetCurrentUser,
		"GetUserClaim":        testGetUserClaim,
		"GetTokenWorkspace":   testGetTokenWorkspace,
		"GetActiveOrg":        testGetActiveOrg,
		"GetDatasetClaim":     testGetDatasetClaim,
		"GetManifest":         testGetManifest,
		"GetOrgClaim":         testGetOrgClaim,
		"GetOrgClaimByNodeId": testGetOrgClaimByNodeId,
		"GetTeamClaims":       testGetTeamClaims,
	} {
		t.Run(scenario, func(t *testing.T) {

			t.Run("token without workspace", func(t *testing.T) {
				noWorkspaceParams := mocks.NewClaimsManagerParams(t)
				tstFunc(t, noWorkspaceParams)
				noWorkspaceParams.AssertMockExpectations(t)
			})

			t.Run("token with workspace", func(t *testing.T) {
				tokenWorkspace := manager.TokenWorkspace{
					Id:     5001,
					NodeId: fmt.Sprintf("N:organization:%s", uuid.NewString()),
				}
				withWorkspaceParams := mocks.NewClaimsManagerParams(t).WithTokenWorkspace(t, tokenWorkspace)
				tstFunc(t, withWorkspaceParams)
				withWorkspaceParams.AssertMockExpectations(t)
			})

		})
	}

}

func testGetCurrentUser(t *testing.T, params *mocks.ClaimsManagerParams) {
	expectedUser := test.NewUser(101, 2001)
	claimsManager := params.WithUserQueryMocked(t, expectedUser).BuildClaimsManager()

	ctx := context.Background()
	user, err := claimsManager.GetCurrentUser(ctx)
	require.NoError(t, err)
	assert.Equal(t, expectedUser, user)
}

func testGetUserClaim(t *testing.T, params *mocks.ClaimsManagerParams) {
	claimsManager := params.BuildClaimsManager()

	expectedUser := test.NewUser(101, 2001)
	ctx := context.Background()

	userClaim := claimsManager.GetUserClaim(ctx, expectedUser)
	assert.Equal(t, expectedUser.Id, userClaim.Id)
	assert.Equal(t, expectedUser.NodeId, userClaim.NodeId)
	assert.Equal(t, expectedUser.IsSuperAdmin, userClaim.IsSuperAdmin)

}

func testGetTokenWorkspace(t *testing.T, params *mocks.ClaimsManagerParams) {
	claimsManager := params.BuildClaimsManager()

	tokenWorkspace, hasTokenWorkspace := claimsManager.GetTokenWorkspace()
	if params.TestJWT.Workspace == nil {
		assert.False(t, hasTokenWorkspace)
	} else {
		assert.True(t, hasTokenWorkspace)
		assert.Equal(t, *params.TestJWT.Workspace, tokenWorkspace)
	}
}

func testGetActiveOrg(t *testing.T, params *mocks.ClaimsManagerParams) {
	claimsManager := params.BuildClaimsManager()

	expectedUser := test.NewUser(101, 2001)
	expectedOrgId := params.GetExpectedOrgId(expectedUser)

	ctx := context.Background()
	orgId := claimsManager.GetActiveOrg(ctx, expectedUser)
	assert.Equal(t, expectedOrgId, orgId)
	if params.TestJWT.Workspace != nil {
		assert.NotEqual(t, expectedUser.PreferredOrg, orgId)
	}
}

func testGetDatasetClaim(t *testing.T, params *mocks.ClaimsManagerParams) {
	claimsManager := params.BuildClaimsManager()

	// Set up Mock
	expectedUser := test.NewUser(101, 2001)
	datasetId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	expectedOrgId := params.GetExpectedOrgId(expectedUser)
	expectedDatasetClaim := &dataset.Claim{
		Role:   role.Viewer,
		NodeId: datasetId,
		IntId:  555,
	}
	params.MockPennsievePg.OnGetDatasetClaim(expectedUser, datasetId, expectedOrgId).Return(expectedDatasetClaim, nil)

	ctx := context.Background()
	claim, err := claimsManager.GetDatasetClaim(ctx, expectedUser, datasetId, expectedOrgId)
	require.NoError(t, err)

	assert.Equal(t, expectedDatasetClaim, claim)
}

func testGetManifest(t *testing.T, params *mocks.ClaimsManagerParams) {
	claimsManager := params.BuildClaimsManager()

	// set up mock
	expectedManifestId := uuid.NewString()
	expectedManifest := &dydb.ManifestTable{
		ManifestId:    expectedManifestId,
		DatasetId:     555,
		DatasetNodeId: fmt.Sprintf("N:dataset:%s", uuid.NewString()),
	}
	params.MockPennsieveDy.OnGetManifestById(params.ManifestTableName, expectedManifestId).Return(expectedManifest, nil)

	ctx := context.Background()
	manifest, err := claimsManager.GetManifest(ctx, expectedManifestId)
	require.NoError(t, err)
	assert.Equal(t, expectedManifest, manifest)
}

func testGetOrgClaim(t *testing.T, params *mocks.ClaimsManagerParams) {
	claimsManager := params.BuildClaimsManager()

	expectedUser := test.NewUser(101, 2001)
	expectedOrgId := params.GetExpectedOrgId(expectedUser)
	expectedOrgNodeId := params.GetExpectedOrgNodeId()
	expectedClaim := &organization.Claim{
		Role:            pgdb.Owner,
		IntId:           expectedOrgId,
		NodeId:          expectedOrgNodeId,
		EnabledFeatures: []pgdb.FeatureFlags{{OrganizationId: expectedOrgId, Feature: "test-feature", Enabled: true}},
	}
	params.MockPennsievePg.OnGetOrganizationClaim(expectedUser.Id, expectedOrgId).Return(expectedClaim, nil)

	ctx := context.Background()
	claim, err := claimsManager.GetOrgClaim(ctx, expectedUser.Id, expectedOrgId)
	require.NoError(t, err)
	assert.Equal(t, expectedClaim, claim)
}

func testGetOrgClaimByNodeId(t *testing.T, params *mocks.ClaimsManagerParams) {
	claimsManager := params.BuildClaimsManager()

	expectedUser := test.NewUser(101, 2001)
	expectedOrgId := params.GetExpectedOrgId(expectedUser)
	expectedOrgNodeId := params.GetExpectedOrgNodeId()
	expectedClaim := &organization.Claim{
		Role:            pgdb.Owner,
		IntId:           expectedOrgId,
		NodeId:          expectedOrgNodeId,
		EnabledFeatures: []pgdb.FeatureFlags{{OrganizationId: expectedOrgId, Feature: "test-feature", Enabled: true}},
	}
	params.MockPennsievePg.OnGetOrganizationClaimByNodeId(expectedUser.Id, expectedOrgNodeId).Return(expectedClaim, nil)

	ctx := context.Background()
	claim, err := claimsManager.GetOrgClaimByNodeId(ctx, expectedUser.Id, expectedOrgNodeId)
	require.NoError(t, err)
	assert.Equal(t, expectedClaim, claim)
}

func testGetTeamClaims(t *testing.T, params *mocks.ClaimsManagerParams) {
	claimsManager := params.BuildClaimsManager()

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
	params.MockPennsievePg.OnGetTeamClaims(expectedUser.Id).Return(expectedClaims, nil)

	ctx := context.Background()
	claims, err := claimsManager.GetTeamClaims(ctx, expectedUser.Id)
	require.NoError(t, err)
	assert.Equal(t, expectedClaims, claims)
}
