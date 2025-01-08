package mocks

import (
	"context"
	"fmt"
	"github.com/google/uuid"

	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/teamUser"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
)

type MockClaimManager struct{}

func NewMockClaimManager() manager.IdentityManager {
	return &MockClaimManager{}
}

func (m *MockClaimManager) GetActiveOrg(context.Context, *pgdbModels.User) int64 {
	return 1
}

func (m *MockClaimManager) GetCurrentUser(context.Context) (*pgdbModels.User, error) {
	return &pgdbModels.User{
		Id:           1,
		NodeId:       "N:user:someRandomUuid",
		IsSuperAdmin: true,
	}, nil
}

var MockUserClaim = user.Claim{
	Id:           1,
	NodeId:       "N:user:someRandomUuid",
	IsSuperAdmin: true,
}

func (m *MockClaimManager) GetUserClaim(context.Context, *pgdbModels.User) *user.Claim {
	return &MockUserClaim
}

var MockDatasetClaim = dataset.Claim{Role: role.Manager}

func (m *MockClaimManager) GetDatasetClaim(context.Context, *pgdbModels.User, string, int64) (*dataset.Claim, error) {
	return &MockDatasetClaim, nil
}

var MockOrgClaim = organization.Claim{}

func (m *MockClaimManager) GetOrgClaim(context.Context, int64, int64) (*organization.Claim, error) {
	return &MockOrgClaim, nil
}

var MockTeamClaims = []teamUser.Claim{{IntId: 1, Name: "someTeam1"}}

func (m *MockClaimManager) GetOrgClaimByNodeId(ctx context.Context, userId int64, orgNodeId string) (*organization.Claim, error) {
	return nil, fmt.Errorf("mock method not implemented")
}

func (m *MockClaimManager) GetTeamClaims(context.Context, int64) ([]teamUser.Claim, error) {
	return MockTeamClaims, nil
}

func (m *MockClaimManager) GetDatasetID(context.Context, string) (string, error) {
	s := "someDatasetID"
	return s, nil
}

func (m *MockClaimManager) GetTokenWorkspace() (manager.TokenWorkspace, bool) {
	return manager.TokenWorkspace{
		Id:     7,
		NodeId: fmt.Sprintf("N:organization:%s", uuid.NewString()),
	}, true
}
