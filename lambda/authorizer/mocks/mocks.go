package mocks

import (
	"context"

	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset/role"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
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

func (m *MockClaimManager) GetUserClaim(context.Context, *pgdbModels.User) user.Claim {
	return user.Claim{
		Id:           1,
		NodeId:       "N:user:someRandomUuid",
		IsSuperAdmin: true,
	}
}

func (m *MockClaimManager) GetDatasetClaim(context.Context, *pgdbModels.User, string, int64) (*dataset.Claim, error) {
	return &dataset.Claim{Role: role.Manager}, nil
}

func (m *MockClaimManager) GetOrgClaim(context.Context, *pgdbModels.User, int64) (*organization.Claim, error) {
	return &organization.Claim{}, nil
}

func (m *MockClaimManager) GetTeamClaims(context.Context, *pgdbModels.User) ([]teamUser.Claim, error) {
	return []teamUser.Claim{}, nil
}
