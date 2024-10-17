package service

import (
	"context"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/pennsieve/pennsieve-go-api/authorizer/mappers"
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
	authFactory := mappers.NewCustomAuthorizerFactory()
	return mappers.IdentitySourceToAuthorizer(i.IdentitySource, authFactory)
}
