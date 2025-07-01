package authorizers_test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test/mocks"
	coreAuthorizer "github.com/pennsieve/pennsieve-go-core/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dydb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/teamUser"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/stretchr/testify/assert"
)

func TestManifestAuthorizer(t *testing.T) {

	for scenario, tstFunc := range map[string]func(t *testing.T, params *mocks.ClaimsManagerParams){
		"GenerateClaims":                        testGenerateClaims,
		"GenerateClaims, Legacy":                testGenerateClaimsLegacy,
		"GenerateClaims, no dataset permission": testGenerateClaimsNoDatasetPermission,
		"GenerateClaims, no org permission":     testGenerateClaimsNoOrgPermission,
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

func testGenerateClaims(t *testing.T, managerParams *mocks.ClaimsManagerParams) {
	//Setup
	currentUser := test.NewUser(101, 1001)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	manifestId := uuid.NewString()
	datasetId := int64(999)
	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	expectedOrgId := managerParams.GetExpectedOrgId(currentUser)
	expectedOrgNodeId := managerParams.GetExpectedOrgNodeId()
	orgClaim := &organization.Claim{
		Role:            pgdb.Read,
		IntId:           expectedOrgId,
		NodeId:          expectedOrgNodeId,
		EnabledFeatures: nil,
	}
	manifest := &dydb.ManifestTable{
		ManifestId:     manifestId,
		DatasetId:      datasetId,
		DatasetNodeId:  datasetNodeId,
		OrganizationId: expectedOrgId,
		UserId:         currentUser.Id,
	}
	datasetClaim := &dataset.Claim{
		Role:   role.Viewer,
		NodeId: datasetNodeId,
		IntId:  datasetId,
	}
	managerParams.MockPennsievePg.OnGetOrganizationClaim(currentUser.Id, expectedOrgId).Return(orgClaim, nil)
	managerParams.MockPennsieveDy.OnGetManifestById(managerParams.ManifestTableName, manifestId).Return(manifest, nil)
	managerParams.MockPennsievePg.OnGetDatasetClaim(currentUser, datasetNodeId, expectedOrgId).Return(datasetClaim, nil)

	// Test
	authorizer := authorizers.NewManifestAuthorizer(manifestId)
	claims, err := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	// Checking results
	require.NoError(t, err)

	assert.Equal(t, 3, len(claims))
	assert.Equal(t,
		expectedUserClaim(currentUser),
		claims[coreAuthorizer.LabelUserClaim])
	assert.Equal(t,
		orgClaim,
		claims[coreAuthorizer.LabelOrganizationClaim])
	assert.Equal(t,
		datasetClaim,
		claims[coreAuthorizer.LabelDatasetClaim])
}

func testGenerateClaimsLegacy(t *testing.T, managerParams *mocks.ClaimsManagerParams) {
	//Setup
	currentUser := test.NewUser(101, 1001)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	manifestId := uuid.NewString()
	datasetId := int64(999)
	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	expectedOrgId := managerParams.GetExpectedOrgId(currentUser)
	expectedOrgNodeId := managerParams.GetExpectedOrgNodeId()
	orgClaim := &organization.Claim{
		Role:            pgdb.Read,
		IntId:           expectedOrgId,
		NodeId:          expectedOrgNodeId,
		EnabledFeatures: nil,
	}
	manifest := &dydb.ManifestTable{
		ManifestId:     manifestId,
		DatasetId:      datasetId,
		DatasetNodeId:  datasetNodeId,
		OrganizationId: expectedOrgId,
		UserId:         currentUser.Id,
	}
	datasetClaim := &dataset.Claim{
		Role:   role.Viewer,
		NodeId: datasetNodeId,
		IntId:  datasetId,
	}
	teamClaims := []teamUser.Claim{
		{
			IntId:      10,
			Name:       "team 1",
			NodeId:     uuid.NewString(),
			Permission: pgdb.Write,
		},
		{
			IntId:      20,
			Name:       "team 2",
			NodeId:     uuid.NewString(),
			Permission: pgdb.Read,
		},
	}
	managerParams.MockPennsievePg.OnGetOrganizationClaim(currentUser.Id, expectedOrgId).Return(orgClaim, nil)
	managerParams.MockPennsieveDy.OnGetManifestById(managerParams.ManifestTableName, manifestId).Return(manifest, nil)
	managerParams.MockPennsievePg.OnGetDatasetClaim(currentUser, datasetNodeId, expectedOrgId).Return(datasetClaim, nil)
	managerParams.MockPennsievePg.OnGetTeamClaims(currentUser.Id).Return(teamClaims, nil)

	// Test
	authorizer := authorizers.NewManifestAuthorizer(manifestId)
	claims, err := authorizer.GenerateClaims(context.Background(), claimsManager, "LEGACY")

	// Checking results
	require.NoError(t, err)

	assert.Equal(t, 4, len(claims))
	assert.Equal(t,
		expectedUserClaim(currentUser),
		claims[coreAuthorizer.LabelUserClaim])
	assert.Equal(t,
		orgClaim,
		claims[coreAuthorizer.LabelOrganizationClaim])
	assert.Equal(t,
		datasetClaim,
		claims[coreAuthorizer.LabelDatasetClaim])
	assert.Equal(t, teamClaims, claims[coreAuthorizer.LabelTeamClaims])
}

func testGenerateClaimsNoDatasetPermission(t *testing.T, managerParams *mocks.ClaimsManagerParams) {
	//Setup
	currentUser := test.NewUser(101, 1001)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	manifestId := uuid.NewString()
	datasetId := int64(999)
	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	expectedOrgId := managerParams.GetExpectedOrgId(currentUser)
	expectedOrgNodeId := managerParams.GetExpectedOrgNodeId()
	orgClaim := &organization.Claim{
		Role:            pgdb.Read,
		IntId:           expectedOrgId,
		NodeId:          expectedOrgNodeId,
		EnabledFeatures: nil,
	}
	manifest := &dydb.ManifestTable{
		ManifestId:     manifestId,
		DatasetId:      datasetId,
		DatasetNodeId:  datasetNodeId,
		OrganizationId: expectedOrgId,
		UserId:         currentUser.Id,
	}
	datasetClaim := &dataset.Claim{
		Role:   role.None, //No dataset role for user, so GenerateClaims should error
		NodeId: datasetNodeId,
		IntId:  datasetId,
	}
	managerParams.MockPennsieveDy.OnGetManifestById(managerParams.ManifestTableName, manifestId).Return(manifest, nil)
	managerParams.MockPennsievePg.OnGetOrganizationClaim(currentUser.Id, expectedOrgId).Return(orgClaim, nil)
	managerParams.MockPennsievePg.OnGetDatasetClaim(currentUser, datasetNodeId, expectedOrgId).Return(datasetClaim, nil)

	// Test
	authorizer := authorizers.NewManifestAuthorizer(manifestId)
	_, err := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	// Checking results
	assert.ErrorContains(t, err, "user has no access to dataset")
}

func testGenerateClaimsNoOrgPermission(t *testing.T, managerParams *mocks.ClaimsManagerParams) {
	//Setup
	currentUser := test.NewUser(101, 1001)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	manifestId := uuid.NewString()
	expectedOrgId := managerParams.GetExpectedOrgId(currentUser)
	expectedOrgNodeId := managerParams.GetExpectedOrgNodeId()
	manifest := &dydb.ManifestTable{
		ManifestId:     manifestId,
		DatasetId:      555,
		DatasetNodeId:  fmt.Sprintf("N:dataset:%s", uuid.NewString()),
		OrganizationId: expectedOrgId,
		UserId:         currentUser.Id,
	}
	orgClaim := &organization.Claim{
		Role:            pgdb.NoPermission, //No org permission for user, so GenerateClaims should error
		IntId:           expectedOrgId,
		NodeId:          expectedOrgNodeId,
		EnabledFeatures: nil,
	}
	managerParams.MockPennsieveDy.OnGetManifestById(managerParams.ManifestTableName, manifestId).Return(manifest, nil)
	managerParams.MockPennsievePg.OnGetOrganizationClaim(currentUser.Id, expectedOrgId).Return(orgClaim, nil)

	// Test
	authorizer := authorizers.NewManifestAuthorizer(manifestId)
	_, err := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	// Checking results
	assert.ErrorContains(t, err, "user has no access to workspace")
}

// TestManifestOrgDoesNotMatchPreferredOrg is not part of main test above, since it only applies to
// non-workspace tokens
func TestManifestOrgDoesNotMatchPreferredOrg(t *testing.T) {
	//Setup
	managerParams := mocks.NewClaimsManagerParams(t)
	userPreferredOrg := int64(1001)
	currentUser := test.NewUser(101, userPreferredOrg)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	// point of test is that this is different from user's PreferredOrgId, which should be ignored
	manifestOrgId := int64(6001)
	manifestOrgNodeId := fmt.Sprintf("N:organization:%s", uuid.NewString())
	manifestId := uuid.NewString()
	datasetId := int64(999)
	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	orgClaim := &organization.Claim{
		Role:            pgdb.Read,
		IntId:           manifestOrgId,
		NodeId:          manifestOrgNodeId,
		EnabledFeatures: nil,
	}
	manifest := &dydb.ManifestTable{
		ManifestId:     manifestId,
		DatasetId:      datasetId,
		DatasetNodeId:  datasetNodeId,
		OrganizationId: manifestOrgId,
		UserId:         currentUser.Id,
	}
	datasetClaim := &dataset.Claim{
		Role:   role.Viewer,
		NodeId: datasetNodeId,
		IntId:  datasetId,
	}
	managerParams.MockPennsievePg.OnGetOrganizationClaim(currentUser.Id, manifestOrgId).Return(orgClaim, nil)
	managerParams.MockPennsieveDy.OnGetManifestById(managerParams.ManifestTableName, manifestId).Return(manifest, nil)
	managerParams.MockPennsievePg.OnGetDatasetClaim(currentUser, datasetNodeId, manifestOrgId).Return(datasetClaim, nil)

	// Test
	authorizer := authorizers.NewManifestAuthorizer(manifestId)
	claims, err := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	// Checking results
	require.NoError(t, err)

	assert.Equal(t, 3, len(claims))
	assert.Equal(t,
		expectedUserClaim(currentUser),
		claims[coreAuthorizer.LabelUserClaim])
	assert.Equal(t,
		orgClaim,
		claims[coreAuthorizer.LabelOrganizationClaim])
	assert.Equal(t,
		datasetClaim,
		claims[coreAuthorizer.LabelDatasetClaim])
}

// TestManifestOrgDoesNotMatchTokenOrg is not part of main test above, since it only applies to
// tokens that have a workspace
func TestManifestOrgDoesNotMatchTokenOrg(t *testing.T) {
	//Setup
	tokenWorkspace := manager.TokenWorkspace{
		Id:     3001,
		NodeId: fmt.Sprintf("N:organization:%s", uuid.NewString()),
	}
	managerParams := mocks.NewClaimsManagerParams(t).WithTokenWorkspace(t, tokenWorkspace)
	userPreferredOrg := int64(1001)
	currentUser := test.NewUser(101, userPreferredOrg)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	// point of test is that this is different from tokenWorkspace.Id, which should cause an error
	manifestOrgId := int64(6001)
	manifestId := uuid.NewString()
	datasetId := int64(999)
	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	manifest := &dydb.ManifestTable{
		ManifestId:     manifestId,
		DatasetId:      datasetId,
		DatasetNodeId:  datasetNodeId,
		OrganizationId: manifestOrgId,
		UserId:         currentUser.Id,
	}
	managerParams.MockPennsieveDy.OnGetManifestById(managerParams.ManifestTableName, manifestId).Return(manifest, nil)

	// Test
	authorizer := authorizers.NewManifestAuthorizer(manifestId)
	_, err := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	// Checking results
	assert.ErrorContains(t, err, fmt.Sprintf("manifest workspace id %d does not match API token workspace id %d", manifestOrgId, tokenWorkspace.Id))

}

func expectedUserClaim(currentUser *pgdb.User) *user.Claim {
	return &user.Claim{
		Id:           currentUser.Id,
		NodeId:       currentUser.NodeId,
		IsSuperAdmin: currentUser.IsSuperAdmin,
	}
}
