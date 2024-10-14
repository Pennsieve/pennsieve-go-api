package mappers

import (
	"log"
	"regexp"

	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	pgdbQueries "github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
)

func IdentitySourceToAuthorizer(IdentitySource []string, currentUser *pgdbModels.User, pddb *pgdbQueries.Queries, token jwt.Token) authorizers.Authorizer {
	if !matches(IdentitySource[0], `Bearer (?P<token>.*)`) {
		log.Fatalln("token expected to be first identity source")
	}

	switch {
	case len(IdentitySource) > 1 && matches(IdentitySource[1], `N:dataset:`):
		return authorizers.NewDatasetAuthorizer(currentUser, pddb, IdentitySource, token)
	case len(IdentitySource) > 1 && matches(IdentitySource[1], `N:manifest:`):
		return authorizers.NewManifestAuthorizer(currentUser, pddb, IdentitySource, token) // will be deprecated
	case len(IdentitySource) > 1 && matches(IdentitySource[1], `N:organization:`):
		return authorizers.NewWorkspaceAuthorizer(currentUser, pddb, IdentitySource, token)
	default:
		return authorizers.NewUserAuthorizer(currentUser)
	}

}

func matches(stringToMatch string, expression string) bool {
	r := regexp.MustCompile(expression)
	parts := r.FindStringSubmatch(stringToMatch)
	return parts != nil
}
