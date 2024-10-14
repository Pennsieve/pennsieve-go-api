package mappers

import (
	"log"
	"regexp"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
)

func IdentitySourceToAuthorizer(IdentitySource []string) authorizers.Authorizer {
	if !matches(IdentitySource[0], `Bearer (?P<token>.*)`) {
		log.Fatalln("token expected to be first identity source")
	}

	switch {
	case len(IdentitySource) > 1 && matches(IdentitySource[1], `N:dataset:`):
		return authorizers.NewDatasetAuthorizer()
	case len(IdentitySource) > 1 && matches(IdentitySource[1], `N:organization:`):
		return authorizers.NewWorkspaceAuthorizer()
	default:
		return authorizers.NewUserAuthorizer()
	}

}

func matches(stringToMatch string, expression string) bool {
	r := regexp.MustCompile(expression)
	parts := r.FindStringSubmatch(stringToMatch)
	return parts != nil
}
