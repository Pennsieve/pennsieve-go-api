package authorizers_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test/mocks"
	coreAuthorizer "github.com/pennsieve/pennsieve-go-core/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/teamUser"
	corePgdb "github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/stretchr/testify/assert"
)

func TestDatasetAuthorizer(t *testing.T) {

	for scenario, tstFunc := range map[string]func(t *testing.T, params *mocks.ClaimsManagerParams){
		"GenerateClaims":                        testDatasetGenerateClaims,
		"GenerateClaims, Legacy":                testDatasetGenerateClaimsLegacy,
		"GenerateClaims, no dataset permission": testDatasetGenerateClaimsNoDatasetPermission,
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

func testDatasetGenerateClaims(t *testing.T, managerParams *mocks.ClaimsManagerParams) {
	//Setup
	currentUser := test.NewUser(101, 1001)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	expectedOrgId := managerParams.GetExpectedOrgId(currentUser)
	expectedOrgNodeId := managerParams.GetExpectedOrgNodeId()
	orgClaim := &organization.Claim{
		Role:            pgdb.Read,
		IntId:           expectedOrgId,
		NodeId:          expectedOrgNodeId,
		EnabledFeatures: nil,
	}
	datasetClaim := &dataset.Claim{
		Role:   role.Viewer,
		NodeId: datasetNodeId,
		IntId:  999,
	}
	managerParams.MockPennsievePg.OnGetOrganizationIdForDataset(datasetNodeId).Return(expectedOrgId, nil)
	managerParams.MockPennsievePg.OnGetOrganizationClaim(currentUser.Id, expectedOrgId).Return(orgClaim, nil)
	managerParams.MockPennsievePg.OnGetDatasetClaim(currentUser, datasetNodeId, expectedOrgId).Return(datasetClaim, nil)

	// Test
	authorizer := authorizers.NewDatasetAuthorizer(datasetNodeId)
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

func testDatasetGenerateClaimsLegacy(t *testing.T, managerParams *mocks.ClaimsManagerParams) {
	//Setup
	currentUser := test.NewUser(101, 1001)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	expectedOrgId := managerParams.GetExpectedOrgId(currentUser)
	expectedOrgNodeId := managerParams.GetExpectedOrgNodeId()
	orgClaim := &organization.Claim{
		Role:            pgdb.Read,
		IntId:           expectedOrgId,
		NodeId:          expectedOrgNodeId,
		EnabledFeatures: nil,
	}
	datasetClaim := &dataset.Claim{
		Role:   role.Viewer,
		NodeId: datasetNodeId,
		IntId:  999,
	}
	teamClaims := []teamUser.Claim{
		{IntId: 10, Name: "team 1", NodeId: uuid.NewString(), Permission: pgdb.Write},
		{IntId: 20, Name: "team 2", NodeId: uuid.NewString(), Permission: pgdb.Read},
	}
	managerParams.MockPennsievePg.OnGetOrganizationIdForDataset(datasetNodeId).Return(expectedOrgId, nil)
	managerParams.MockPennsievePg.OnGetOrganizationClaim(currentUser.Id, expectedOrgId).Return(orgClaim, nil)
	managerParams.MockPennsievePg.OnGetDatasetClaim(currentUser, datasetNodeId, expectedOrgId).Return(datasetClaim, nil)
	managerParams.MockPennsievePg.OnGetTeamClaimsForOrg(currentUser.Id, expectedOrgId).Return(teamClaims, nil)

	// Test
	authorizer := authorizers.NewDatasetAuthorizer(datasetNodeId)
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

func testDatasetGenerateClaimsNoDatasetPermission(t *testing.T, managerParams *mocks.ClaimsManagerParams) {
	//Setup
	currentUser := test.NewUser(101, 1001)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	expectedOrgId := managerParams.GetExpectedOrgId(currentUser)
	expectedOrgNodeId := managerParams.GetExpectedOrgNodeId()
	orgClaim := &organization.Claim{
		Role:            pgdb.Read,
		IntId:           expectedOrgId,
		NodeId:          expectedOrgNodeId,
		EnabledFeatures: nil,
	}
	datasetClaim := &dataset.Claim{
		Role:   role.None, //No dataset role for user, so GenerateClaims should error
		NodeId: datasetNodeId,
		IntId:  999,
	}
	managerParams.MockPennsievePg.OnGetOrganizationIdForDataset(datasetNodeId).Return(expectedOrgId, nil)
	managerParams.MockPennsievePg.OnGetOrganizationClaim(currentUser.Id, expectedOrgId).Return(orgClaim, nil)
	managerParams.MockPennsievePg.OnGetDatasetClaim(currentUser, datasetNodeId, expectedOrgId).Return(datasetClaim, nil)

	// Test
	authorizer := authorizers.NewDatasetAuthorizer(datasetNodeId)
	_, err := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	// Checking results: a clean, authoritative (cacheable) deny, not indeterminate.
	assert.ErrorContains(t, err, "user has no access to dataset")
	assertNotIndeterminate(t, err)
}

// TestDatasetOrgDoesNotMatchPreferredOrg is the regression test for the bug this ticket fixes:
// the dataset's actual org differs from the user's preferred org, and that must be ignored
// entirely now that org is resolved from the dataset_organization map, not the user's state.
func TestDatasetOrgDoesNotMatchPreferredOrg(t *testing.T) {
	//Setup
	managerParams := mocks.NewClaimsManagerParams(t)
	userPreferredOrg := int64(1001)
	currentUser := test.NewUser(101, userPreferredOrg)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	// point of test is that this is different from the user's PreferredOrg, which should be ignored
	datasetOrgId := int64(6001)
	datasetOrgNodeId := fmt.Sprintf("N:organization:%s", uuid.NewString())
	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	orgClaim := &organization.Claim{
		Role:            pgdb.Read,
		IntId:           datasetOrgId,
		NodeId:          datasetOrgNodeId,
		EnabledFeatures: nil,
	}
	datasetClaim := &dataset.Claim{
		Role:   role.Viewer,
		NodeId: datasetNodeId,
		IntId:  999,
	}
	managerParams.MockPennsievePg.OnGetOrganizationIdForDataset(datasetNodeId).Return(datasetOrgId, nil)
	managerParams.MockPennsievePg.OnGetOrganizationClaim(currentUser.Id, datasetOrgId).Return(orgClaim, nil)
	managerParams.MockPennsievePg.OnGetDatasetClaim(currentUser, datasetNodeId, datasetOrgId).Return(datasetClaim, nil)

	// Test
	authorizer := authorizers.NewDatasetAuthorizer(datasetNodeId)
	claims, err := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	// Checking results
	require.NoError(t, err)
	assert.Equal(t, 3, len(claims))
	assert.Equal(t, orgClaim, claims[coreAuthorizer.LabelOrganizationClaim])
	assert.Equal(t, datasetClaim, claims[coreAuthorizer.LabelDatasetClaim])
	managerParams.AssertMockExpectations(t)
}

// TestDatasetOrgDoesNotMatchTokenOrg covers the token-pool (API-key) side: an API key scoped to
// one org must not reach a dataset in another org, even if the underlying user has access there.
func TestDatasetOrgDoesNotMatchTokenOrg(t *testing.T) {
	tokenWorkspace := manager.TokenWorkspace{
		Id:     3001,
		NodeId: fmt.Sprintf("N:organization:%s", uuid.NewString()),
	}
	managerParams := mocks.NewClaimsManagerParams(t).WithTokenWorkspace(t, tokenWorkspace)
	userPreferredOrg := int64(1001)
	currentUser := test.NewUser(101, userPreferredOrg)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	// point of test is that this is different from tokenWorkspace.Id, which should cause a deny
	datasetOrgId := int64(6001)
	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	managerParams.MockPennsievePg.OnGetOrganizationIdForDataset(datasetNodeId).Return(datasetOrgId, nil)

	// Test
	authorizer := authorizers.NewDatasetAuthorizer(datasetNodeId)
	claims, err := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	// Checking results: authoritative, cacheable deny
	assert.Nil(t, claims)
	assert.ErrorContains(t, err,
		fmt.Sprintf("token workspace %d does not match organization %d", tokenWorkspace.Id, datasetOrgId))
	assertNotIndeterminate(t, err)
	managerParams.AssertMockExpectations(t)
}

// TestDatasetOrganizationMapMiss: the dataset genuinely doesn't exist in the dataset_organization
// map. This must be a clean, cacheable deny, not an indeterminate failure.
func TestDatasetOrganizationMapMiss(t *testing.T) {
	managerParams := mocks.NewClaimsManagerParams(t)
	currentUser := test.NewUser(101, 1001)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	managerParams.MockPennsievePg.OnGetOrganizationIdForDataset(datasetNodeId).
		Return(int64(0), corePgdb.DatasetOrganizationNotFoundError{DatasetNodeId: datasetNodeId})

	authorizer := authorizers.NewDatasetAuthorizer(datasetNodeId)
	claims, err := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	assert.Nil(t, claims)
	require.Error(t, err)
	assertNotIndeterminate(t, err)
	managerParams.AssertMockExpectations(t)
}

// TestDatasetOrganizationLookupDBError: a DB/connection failure resolving the dataset's org from
// the map is indeterminate — it must not be cached as a deny.
func TestDatasetOrganizationLookupDBError(t *testing.T) {
	managerParams := mocks.NewClaimsManagerParams(t)
	currentUser := test.NewUser(101, 1001)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	managerParams.MockPennsievePg.OnGetOrganizationIdForDataset(datasetNodeId).
		Return(int64(0), errors.New("connection refused"))

	authorizer := authorizers.NewDatasetAuthorizer(datasetNodeId)
	claims, err := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	assert.Nil(t, claims)
	require.Error(t, err)
	assertIndeterminate(t, err)
	managerParams.AssertMockExpectations(t)
}

// TestDatasetOrgClaimDBError: a DB failure fetching the org role is also indeterminate.
func TestDatasetOrgClaimDBError(t *testing.T) {
	managerParams := mocks.NewClaimsManagerParams(t)
	currentUser := test.NewUser(101, 1001)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	expectedOrgId := managerParams.GetExpectedOrgId(currentUser)
	managerParams.MockPennsievePg.OnGetOrganizationIdForDataset(datasetNodeId).Return(expectedOrgId, nil)
	managerParams.MockPennsievePg.OnGetOrganizationClaim(currentUser.Id, expectedOrgId).
		Return((*organization.Claim)(nil), errors.New("connection refused"))

	authorizer := authorizers.NewDatasetAuthorizer(datasetNodeId)
	claims, err := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	assert.Nil(t, claims)
	require.Error(t, err)
	assertIndeterminate(t, err)
	managerParams.AssertMockExpectations(t)
}

// TestDatasetUserNotOrgMember: GetOrganizationClaim's inner join returns
// OrganizationUserNotFoundError (not a NoPermission claim) when the user has no
// organization_user row for the resolved org. This must be a clean, cacheable deny,
// not indeterminate — the user genuinely isn't a member of that org.
func TestDatasetUserNotOrgMember(t *testing.T) {
	managerParams := mocks.NewClaimsManagerParams(t)
	currentUser := test.NewUser(101, 1001)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	expectedOrgId := managerParams.GetExpectedOrgId(currentUser)
	managerParams.MockPennsievePg.OnGetOrganizationIdForDataset(datasetNodeId).Return(expectedOrgId, nil)
	managerParams.MockPennsievePg.OnGetOrganizationClaim(currentUser.Id, expectedOrgId).
		Return((*organization.Claim)(nil), corePgdb.OrganizationUserNotFoundError{ErrorMessage: "no rows"})

	authorizer := authorizers.NewDatasetAuthorizer(datasetNodeId)
	claims, err := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	assert.Nil(t, claims)
	require.Error(t, err)
	assertNotIndeterminate(t, err)
	managerParams.AssertMockExpectations(t)
}

// TestDatasetClaimDBError: a DB failure fetching the dataset role is also indeterminate.
func TestDatasetClaimDBError(t *testing.T) {
	managerParams := mocks.NewClaimsManagerParams(t)
	currentUser := test.NewUser(101, 1001)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	expectedOrgId := managerParams.GetExpectedOrgId(currentUser)
	expectedOrgNodeId := managerParams.GetExpectedOrgNodeId()
	orgClaim := &organization.Claim{
		Role:            pgdb.Read,
		IntId:           expectedOrgId,
		NodeId:          expectedOrgNodeId,
		EnabledFeatures: nil,
	}
	managerParams.MockPennsievePg.OnGetOrganizationIdForDataset(datasetNodeId).Return(expectedOrgId, nil)
	managerParams.MockPennsievePg.OnGetOrganizationClaim(currentUser.Id, expectedOrgId).Return(orgClaim, nil)
	managerParams.MockPennsievePg.OnGetDatasetClaim(currentUser, datasetNodeId, expectedOrgId).
		Return((*dataset.Claim)(nil), errors.New("connection refused"))

	authorizer := authorizers.NewDatasetAuthorizer(datasetNodeId)
	claims, err := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	assert.Nil(t, claims)
	require.Error(t, err)
	assertIndeterminate(t, err)
	managerParams.AssertMockExpectations(t)
}

// TestDatasetTeamClaimsForOrgDBError: a DB failure fetching (LEGACY mode) team claims is also indeterminate.
func TestDatasetTeamClaimsForOrgDBError(t *testing.T) {
	managerParams := mocks.NewClaimsManagerParams(t)
	currentUser := test.NewUser(101, 1001)
	claimsManager := managerParams.WithUserQueryMocked(t, currentUser).BuildClaimsManager()

	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	expectedOrgId := managerParams.GetExpectedOrgId(currentUser)
	expectedOrgNodeId := managerParams.GetExpectedOrgNodeId()
	orgClaim := &organization.Claim{
		Role:            pgdb.Read,
		IntId:           expectedOrgId,
		NodeId:          expectedOrgNodeId,
		EnabledFeatures: nil,
	}
	datasetClaim := &dataset.Claim{
		Role:   role.Viewer,
		NodeId: datasetNodeId,
		IntId:  999,
	}
	managerParams.MockPennsievePg.OnGetOrganizationIdForDataset(datasetNodeId).Return(expectedOrgId, nil)
	managerParams.MockPennsievePg.OnGetOrganizationClaim(currentUser.Id, expectedOrgId).Return(orgClaim, nil)
	managerParams.MockPennsievePg.OnGetDatasetClaim(currentUser, datasetNodeId, expectedOrgId).Return(datasetClaim, nil)
	managerParams.MockPennsievePg.OnGetTeamClaimsForOrg(currentUser.Id, expectedOrgId).
		Return([]teamUser.Claim(nil), errors.New("connection refused"))

	authorizer := authorizers.NewDatasetAuthorizer(datasetNodeId)
	claims, err := authorizer.GenerateClaims(context.Background(), claimsManager, "LEGACY")

	assert.Nil(t, claims)
	require.Error(t, err)
	assertIndeterminate(t, err)
	managerParams.AssertMockExpectations(t)
}

func assertIndeterminate(t *testing.T, err error) {
	t.Helper()
	var indeterminate *authorizers.IndeterminateError
	assert.True(t, errors.As(err, &indeterminate), "expected an indeterminate error (uncached 500), got: %v", err)
}

func assertNotIndeterminate(t *testing.T, err error) {
	t.Helper()
	var indeterminate *authorizers.IndeterminateError
	assert.False(t, errors.As(err, &indeterminate), "expected an authoritative deny (not indeterminate), got: %v", err)
}
