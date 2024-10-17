package service

import (
	"context"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	factory "github.com/pennsieve/pennsieve-go-api/authorizer/factory"
)

type IdentityService interface {
	GetAuthorizer(context.Context) (authorizers.Authorizer, error)
}

type IdentitySourceService struct {
	IdentitySource []string
}

func NewIdentitySourceService(IdentitySource []string) IdentityService {
	return &IdentitySourceService{IdentitySource}
}

func (i *IdentitySourceService) GetAuthorizer(ctx context.Context) (authorizers.Authorizer, error) {
	authFactory := factory.NewCustomAuthorizerFactory()
	return authFactory.Build(i.IdentitySource)
}
