package helpers

import (
	"errors"
	"net/url"
	"regexp"
)

func Matches(stringToMatch string, expression string) bool {
	r := regexp.MustCompile(expression)
	parts := r.FindStringSubmatch(stringToMatch)
	return parts != nil
}

func DecodeIdentitySource(identitySource string) (string, error) {
	return url.QueryUnescape(identitySource)
}

func GetJWT(authorization string) ([]byte, error) {
	r := regexp.MustCompile(`Bearer (?P<token>.*)`)
	tokenParts := r.FindStringSubmatch(authorization)
	if len(tokenParts) == 0 {
		return nil, errors.New("expected token to be in the format: Bearer <token>")
	}

	return []byte(tokenParts[r.SubexpIndex("token")]), nil
}
