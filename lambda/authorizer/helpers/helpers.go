package helpers

import (
	"errors"
	"regexp"
)

func Matches(stringToMatch string, expression string) bool {
	r := regexp.MustCompile(expression)
	parts := r.FindStringSubmatch(stringToMatch)
	return parts != nil
}

func GetJWT(authorization string) ([]byte, error) {
	r := regexp.MustCompile(`Bearer (?P<token>.*)`)
	tokenParts := r.FindStringSubmatch(authorization)
	if len(tokenParts) == 0 {
		return nil, errors.New("expected token to be in the format: Bearer <token>")
	}

	return []byte(tokenParts[r.SubexpIndex("token")]), nil
}
