package test

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/stretchr/testify/require"
	"math/rand"
)

// JWT wraps a jwt.Token. Create one with either NewJWT, NewJWTWithWorkspace, or JWTBuilder so
// that the claim values are automatically written to JWT.Username, JWT.ClientId, and optionally JWT.Workspace as fields for easy access in tests
type JWT struct {
	Username  string
	ClientId  string
	Workspace *manager.TokenWorkspace
	Token     jwt.Token
}

func NewJWT(t require.TestingT) JWT {
	return NewJWTBuilder().Build(t)
}

func NewJWTWithWorkspace(t require.TestingT) JWT {
	return NewJWTBuilder().WithRandomWorkspace().Build(t)
}

type JWTBuilder struct {
	Username  string
	ClientId  string
	Workspace *manager.TokenWorkspace
}

func NewJWTBuilder() *JWTBuilder {
	return &JWTBuilder{}
}

func (b *JWTBuilder) WithUsername(username string) *JWTBuilder {
	b.Username = username
	return b
}

func (b *JWTBuilder) WithClientId(clientId string) *JWTBuilder {
	b.ClientId = clientId
	return b
}

func (b *JWTBuilder) WithRandomWorkspace() *JWTBuilder {
	b.Workspace = &manager.TokenWorkspace{
		Id:     rand.Int63n(100),
		NodeId: fmt.Sprintf("N:organization:%s", uuid.NewString()),
	}
	return b
}

func (b *JWTBuilder) WithWorkspace(id int64, nodeId string) *JWTBuilder {
	b.Workspace = &manager.TokenWorkspace{
		Id:     id,
		NodeId: nodeId,
	}
	return b
}

func (b *JWTBuilder) Build(t require.TestingT) JWT {
	testJWT := JWT{}
	tokenBuilder := jwt.NewBuilder()

	//Username
	if len(b.Username) == 0 {
		testJWT.Username = uuid.NewString()
	} else {
		testJWT.Username = b.Username
	}
	tokenBuilder = tokenBuilder.Claim("username", testJWT.Username)

	//ClientId
	if len(b.ClientId) == 0 {
		testJWT.ClientId = uuid.NewString()
	} else {
		testJWT.ClientId = b.ClientId
	}
	tokenBuilder = tokenBuilder.Claim("client_id", testJWT.ClientId)

	//Workspace
	if b.Workspace != nil {
		testJWT.Workspace = b.Workspace
		if testJWT.Workspace.Id == int64(0) {
			testJWT.Workspace.Id = rand.Int63n(100)
		}
		if len(testJWT.Workspace.NodeId) == 0 {
			testJWT.Workspace.NodeId = fmt.Sprintf("N:organization:%s", uuid.NewString())
		}
		tokenBuilder = tokenBuilder.
			Claim("custom:organization_id", testJWT.Workspace.Id).
			Claim("custom:organization_node_id", testJWT.Workspace.NodeId)
	}
	token, err := tokenBuilder.Build()
	require.NoError(t, err)
	testJWT.Token = token
	return testJWT
}
