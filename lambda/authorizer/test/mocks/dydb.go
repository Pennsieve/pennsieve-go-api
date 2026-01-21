package mocks

import (
	"context"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dydb"
	"github.com/stretchr/testify/mock"
)

// MockPennsieveDyAPI is a testify mock of manager.PennsieveDyAPI
// Set up expectations by calling the MockPennsieveDyAPI.On* methods as needed.
// Verify by calling MockPennsieveDyAPI.AssertExpectations(t)
type MockPennsieveDyAPI struct {
	mock.Mock
}

func NewMockPennsieveDyAPI() *MockPennsieveDyAPI {
	return new(MockPennsieveDyAPI)
}

// manager.PennsieveDyAPI methods

func (m *MockPennsieveDyAPI) GetManifestById(ctx context.Context, manifestTableName string, manifestId string) (*dydb.ManifestTable, error) {
	args := m.Called(ctx, manifestTableName, manifestId)
	return args.Get(0).(*dydb.ManifestTable), args.Error(1)
}

// Helper methods for use in tests to set up expectations for any context.Context value

func (m *MockPennsieveDyAPI) OnGetManifestById(manifestTableName string, manifestId string) *mock.Call {
	return m.On("GetManifestById", mock.Anything, manifestTableName, manifestId)
}
