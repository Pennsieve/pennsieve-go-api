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
	IdentitySource        []string
	QueryStringParameters map[string]string
}

func NewIdentitySourceService(IdentitySource []string, queryStringParameters map[string]string) IdentityService {
	return &IdentitySourceService{IdentitySource, queryStringParameters}
}

func (i *IdentitySourceService) GetAuthorizer(ctx context.Context) (authorizers.Authorizer, error) {
	return factory.NewCustomAuthorizerFactory().Build(i.IdentitySource, i.QueryStringParameters)
}
