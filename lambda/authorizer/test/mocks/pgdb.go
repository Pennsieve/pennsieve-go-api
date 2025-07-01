package mocks

import (
	"context"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/teamUser"
	"github.com/stretchr/testify/mock"
)

// MockPennsievePgAPI is a testify mock of manager.PennsievePgAPI
// Set up expectations by calling the MockPennsievePgAPI.On* methods as needed.
// Verify by calling MockPennsievePgAPI.AssertExpectations(t)
type MockPennsievePgAPI struct {
	mock.Mock
}

func NewMockPennsievePgAPI() *MockPennsievePgAPI {
	return new(MockPennsievePgAPI)
}

// Interface methods for manager.PennsievePgAPI

func (m *MockPennsievePgAPI) GetDatasetClaim(ctx context.Context, user *pgdb.User, datasetNodeId string, organizationId int64) (*dataset.Claim, error) {
	args := m.Called(ctx, user, datasetNodeId, organizationId)
	return args.Get(0).(*dataset.Claim), args.Error(1)
}

func (m *MockPennsievePgAPI) GetOrganizationClaim(ctx context.Context, userId int64, organizationId int64) (*organization.Claim, error) {
	args := m.Called(ctx, userId, organizationId)
	return args.Get(0).(*organization.Claim), args.Error(1)
}

func (m *MockPennsievePgAPI) GetOrganizationClaimByNodeId(ctx context.Context, userId int64, organizationNodeId string) (*organization.Claim, error) {
	args := m.Called(ctx, userId, organizationNodeId)
	return args.Get(0).(*organization.Claim), args.Error(1)
}

func (m *MockPennsievePgAPI) GetTeamClaims(ctx context.Context, userId int64) ([]teamUser.Claim, error) {
	args := m.Called(ctx, userId)
	return args.Get(0).([]teamUser.Claim), args.Error(1)
}

func (m *MockPennsievePgAPI) GetUserByCognitoId(ctx context.Context, cognitoId string) (*pgdb.User, error) {
	args := m.Called(ctx, cognitoId)
	return args.Get(0).(*pgdb.User), args.Error(1)
}

func (m *MockPennsievePgAPI) GetByCognitoId(ctx context.Context, cognitoId string) (*pgdb.User, error) {
	args := m.Called(ctx, cognitoId)
	return args.Get(0).(*pgdb.User), args.Error(1)
}

// Helper methods for use in tests to set up expectations for any context.Context value

func (m *MockPennsievePgAPI) OnGetDatasetClaim(user *pgdb.User, datasetNodeId string, organizationId int64) *mock.Call {
	return m.On("GetDatasetClaim", mock.Anything, user, datasetNodeId, organizationId)
}

func (m *MockPennsievePgAPI) OnGetOrganizationClaim(userId int64, organizationId int64) *mock.Call {
	return m.On("GetOrganizationClaim", mock.Anything, userId, organizationId)
}

func (m *MockPennsievePgAPI) OnGetOrganizationClaimByNodeId(userId int64, organizationNodeId string) *mock.Call {
	return m.On("GetOrganizationClaimByNodeId", mock.Anything, userId, organizationNodeId)
}

func (m *MockPennsievePgAPI) OnGetTeamClaims(userId int64) *mock.Call {
	return m.On("GetTeamClaims", mock.Anything, userId)
}

func (m *MockPennsievePgAPI) OnGetUserByCognitoId(cognitoId string) *mock.Call {
	return m.On("GetUserByCognitoId", mock.Anything, cognitoId)
}

func (m *MockPennsievePgAPI) OnGetByCognitoId(cognitoId string) *mock.Call {
	return m.On("GetByCognitoId", mock.Anything, cognitoId)
}
