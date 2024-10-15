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

func IdentitySourceToAuthorizer(identitySource []string, currentUser *pgdbModels.User, qPgDb *pgdbQueries.Queries, token jwt.Token) (authorizers.Authorizer, error) {
	if !helpers.Matches(identitySource[0], `Bearer (?P<token>.*)`) {
		errorString := "token expected to be first identity source"
		log.Error(errorString)
		return nil, errors.New(errorString)
	}

	switch {
	case len(identitySource) > 1 && helpers.Matches(identitySource[1], `N:dataset:`):
		return authorizers.NewDatasetAuthorizer(currentUser, qPgDb, identitySource, token), nil
	case len(identitySource) > 1 && helpers.Matches(identitySource[1], `N:manifest:`):
		return authorizers.NewManifestAuthorizer(currentUser, qPgDb, identitySource, token), nil // will be deprecated
	case len(identitySource) > 1 && helpers.Matches(identitySource[1], `N:organization:`):
		return authorizers.NewWorkspaceAuthorizer(currentUser, qPgDb, identitySource, token), nil
	default:
		return authorizers.NewUserAuthorizer(currentUser), nil
	}

}
