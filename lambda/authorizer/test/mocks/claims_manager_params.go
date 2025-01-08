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

// ClaimsManagerParams are encapsulates the parameters needed to call manager.NewClaimsManager() for ease of testing.
// The manager.PennsievePgAPI and manager.PennsieveDyAPI parameters are *MockPennsievePgAPI and *MockPennsieveDyAPI
// respectively.
//
// Idea is that if a test requires a manager.ClaimsManager with mock manager.PennsievePgAPI and manager.PennsieveDyAPI
// you can call
//
//	manager := NewClaimsManagerParams(t).BuildClaimsManager()
//
// to get a ClaimsManager with a token that does not contain organization info.
// If you want the manager's JWT token to contain a given tokenWorkspace use
//
//	manager := NewClaimsManagerParams(t).WithTokenWorkspace(t, tokenWorkspace).BuildClaimsManager()
//
// If you have a *pgdb.User that you want MockPennsievePg to return with the correct Get*CognitoId method,
// you can either do
//
//	manager := NewClaimsManagerParams(t).WithUserQueryMocked(t, user).BuildClaimsManager()
//
// which will set up the mock's GetByCognitoId() method or
//
//	manager := NewClaimsManagerParams(t).WithTokenWorkspace(t, tokenWorkspace).WithUserQueryMocked(t, user).BuildClaimsManager()
//
// which will set up the mock's GetUserByCognitoId() method
type ClaimsManagerParams struct {
	MockPennsievePg   *MockPennsievePgAPI
	MockPennsieveDy   *MockPennsieveDyAPI
	TestJWT           test.JWT
	TokenClientId     string
	ManifestTableName string
}

// NewClaimsManagerParams returns a *ClaimsManagerParams with new MockPennsievePgAPI and MockPennsieveDyAPI fields
// and random TokenClientId and ManifestTableName. The TestJWT field will contain a random test.JWT without an organization
func NewClaimsManagerParams(t require.TestingT) *ClaimsManagerParams {
	// A JWT token with no workspace and a clientId
	// different from the tokenClientId (initialized below)
	testJWT := test.NewJWTBuilder().Build(t)

	return &ClaimsManagerParams{
		MockPennsievePg: NewMockPennsievePgAPI(),
		MockPennsieveDy: NewMockPennsieveDyAPI(),
		TestJWT:         testJWT,
		// TokenClientID will be different from the random clientId in TestJWT
		TokenClientId:     uuid.NewString(),
		ManifestTableName: uuid.NewString(),
	}
}

// WithTokenWorkspace updates this *ClaimsManagerParams by replacing the TestJWT field with a new test.JWT instance that contains
// the given manager.TokenWorkspace as its organization. The TokenClientId field of the *ClaimsManagerParams is also updated to match
// the new test.JWT ClientId field.
// The modified *ClaimsManagerParams is returned.
func (p *ClaimsManagerParams) WithTokenWorkspace(t require.TestingT, tokenWorkspace manager.TokenWorkspace) *ClaimsManagerParams {
	// Overwrite TestJWT with a new JWT token with a workspace and a clientId
	// that will match the tokenClientId (initialized below)
	p.TestJWT = test.NewJWTBuilder().
		WithWorkspace(tokenWorkspace.Id, tokenWorkspace.NodeId).
		Build(t)
	p.TokenClientId = p.TestJWT.ClientId

	return p
}

func (p *ClaimsManagerParams) WithUserQueryMocked(t require.TestingT, currentUser *pgdb.User) *ClaimsManagerParams {
	if p.TestJWT.Workspace == nil && p.TokenClientId != p.TestJWT.ClientId {
		// If the jwt does not contain a workspace and did not come from the token pool, then we
		// expect the manager.ClaimsManager to call the pgdb method that queries only the user table
		p.MockPennsievePg.OnGetByCognitoId(p.TestJWT.Username).Return(currentUser, nil)
	} else if p.TestJWT.Workspace != nil && p.TokenClientId == p.TestJWT.ClientId {
		// If the jwt contains a workspace and came from the token pool, then we
		// expect the manager.ClaimsManager to call the pgdb method that queries a join of the users and token tables.
		p.MockPennsievePg.OnGetUserByCognitoId(p.TestJWT.Username).Return(currentUser, nil)
	} else {
		require.FailNow(t, "inconsistent ClaimsManagerParams", "TestJWT should be non-nil if and only if TestJWT clientId is the TokenClientId")
	}
	return p
}

func (p *ClaimsManagerParams) BuildClaimsManager() manager.IdentityManager {
	return manager.NewClaimsManager(p.MockPennsievePg, p.MockPennsieveDy, p.TestJWT.Token, p.TokenClientId, p.ManifestTableName)
}

func (p *ClaimsManagerParams) GetExpectedOrgId(user *pgdb.User) int64 {
	if p.TestJWT.Workspace == nil {
		return user.PreferredOrg
	}
	return p.TestJWT.Workspace.Id
}

func (p *ClaimsManagerParams) GetExpectedOrgNodeId() string {
	if p.TestJWT.Workspace == nil {
		return fmt.Sprintf("N:organization:%s", uuid.NewString())
	}
	return p.TestJWT.Workspace.NodeId
}

func (p *ClaimsManagerParams) AssertMockExpectations(t mock.TestingT) {
	p.MockPennsievePg.AssertExpectations(t)
	p.MockPennsieveDy.AssertExpectations(t)
}
