package helpers

import (
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
