package mappers

import (
	"errors"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/pennsieve/pennsieve-go-api/authorizer/helpers"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	pgdbQueries "github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
)

func IdentitySourceToAuthorizer(identitySource []string, f AuthorizerFactory) (authorizers.Authorizer, error) {
	if !helpers.Matches(identitySource[0], `Bearer (?P<token>.*)`) {
		errorString := "token expected to be first identity source"
		log.Error(errorString)
		return nil, errors.New(errorString)
	}

	return f.Build(identitySource)
}

type AuthorizerFactory interface {
	Build([]string) (authorizers.Authorizer, error)
}

type CustomAuthorizerFactory struct {
	CurrentUser *pgdbModels.User
	Queries     *pgdbQueries.Queries
	Token       jwt.Token
}

func NewCustomAuthorizerFactory(currentUser *pgdbModels.User, qPgDb *pgdbQueries.Queries, token jwt.Token) AuthorizerFactory {
	return &CustomAuthorizerFactory{currentUser, qPgDb, token}
}

func (f *CustomAuthorizerFactory) Build(identitySource []string) (authorizers.Authorizer, error) {
	switch {
	case len(identitySource) > 1 && helpers.Matches(identitySource[1], `N:dataset:`):
		return authorizers.NewDatasetAuthorizer(f.CurrentUser, f.Queries, identitySource, f.Token), nil
	case len(identitySource) > 1 && helpers.Matches(identitySource[1], `N:manifest:`):
		return authorizers.NewManifestAuthorizer(f.CurrentUser, f.Queries, identitySource, f.Token), nil // will be deprecated
	case len(identitySource) > 1 && helpers.Matches(identitySource[1], `N:organization:`):
		return authorizers.NewWorkspaceAuthorizer(f.CurrentUser, f.Queries, identitySource, f.Token), nil
	default:
		return authorizers.NewUserAuthorizer(f.CurrentUser), nil
	}
}
