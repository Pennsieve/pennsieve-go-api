package mocks

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type ManagerParams struct {
	MockPennsievePg   *MockPennsievePgAPI
	MockPennsieveDy   *MockPennsieveDyAPI
	TestJWT           test.JWT
	TokenClientId     string
	ManifestTableName string
}

func (p *ManagerParams) BuildManager() manager.IdentityManager {
	return manager.NewClaimsManager(p.MockPennsievePg, p.MockPennsieveDy, p.TestJWT.Token, p.TokenClientId, p.ManifestTableName)
}

func (p *ManagerParams) WithUserQueryMocked(t require.TestingT, currentUser *pgdb.User) *ManagerParams {
	if p.TestJWT.Workspace == nil && p.TokenClientId != p.TestJWT.ClientId {
		// If the jwt does not contain a workspace and did not come from the token pool, then we
		// expect the manager.ClaimsManager to call the pgdb method that queries only the user table
		p.MockPennsievePg.OnGetByCognitoId(p.TestJWT.Username).Return(currentUser, nil)
	} else if p.TestJWT.Workspace != nil && p.TokenClientId == p.TestJWT.ClientId {
		// If the jwt contains a workspace and came from the token pool, then we
		// expect the manager.ClaimsManager to call the pgdb method that queries a join of the users and token tables.
		p.MockPennsievePg.OnGetUserByCognitoId(p.TestJWT.Username).Return(currentUser, nil)
	} else {
		require.FailNow(t, "inconsistent ManagerParams", "TestJWT should be non-nil if and only if TestJWT clientId is the TokenClientId")
	}
	return p
}

func (p *ManagerParams) GetExpectedOrgId(user *pgdb.User) int64 {
	if p.TestJWT.Workspace == nil {
		return user.PreferredOrg
	}
	return p.TestJWT.Workspace.Id
}

func (p *ManagerParams) GetExpectedOrgNodeId() string {
	if p.TestJWT.Workspace == nil {
		return fmt.Sprintf("N:organization:%s", uuid.NewString())
	}
	return p.TestJWT.Workspace.NodeId
}

func (p *ManagerParams) AssertMockExpectations(t mock.TestingT) {
	p.MockPennsievePg.AssertExpectations(t)
	p.MockPennsieveDy.AssertExpectations(t)
}

func NewNoWorkspaceTokenManagerParams(t require.TestingT) *ManagerParams {
	// A JWT token with no workspace and a clientId
	// different from the tokenClientId (initialized below)
	testJWT := test.NewJWTBuilder().Build(t)

	return &ManagerParams{
		MockPennsievePg: NewMockPennsievePgAPI(),
		MockPennsieveDy: NewMockPennsieveDyAPI(),
		TestJWT:         testJWT,
		// TokenClientID will be different from the random clientId in TestJWT
		TokenClientId:     uuid.NewString(),
		ManifestTableName: uuid.NewString(),
	}
}

func NewWorkspaceTokenManagerParams(t require.TestingT, tokenWorkspace manager.TokenWorkspace) *ManagerParams {
	// A JWT token with a workspace and a clientId
	// that will match the tokenClientId (initialized below)
	testJWT := test.NewJWTBuilder().
		WithWorkspace(tokenWorkspace.Id, tokenWorkspace.NodeId).
		Build(t)

	return &ManagerParams{
		MockPennsievePg: NewMockPennsievePgAPI(),
		MockPennsieveDy: NewMockPennsieveDyAPI(),
		TestJWT:         testJWT,
		// TokenClientID matches the clientId in token
		TokenClientId:     testJWT.ClientId,
		ManifestTableName: uuid.NewString(),
	}
}
