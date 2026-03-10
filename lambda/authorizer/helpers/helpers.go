package helpers

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
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

// CallbackAuth holds the parsed components of a Callback authorization header.
type CallbackAuth struct {
	Service        string
	ExecutionRunID string
	Token          string
}

// ParseCallbackAuth parses an Authorization header of the form:
// Callback <service>:<executionRunId>:<token>
func ParseCallbackAuth(authorization string) (*CallbackAuth, error) {
	if !strings.HasPrefix(authorization, "Callback ") {
		return nil, fmt.Errorf("expected token to be in the format: Callback <service>:<executionRunId>:<token>")
	}

	payload := strings.TrimPrefix(authorization, "Callback ")
	parts := strings.SplitN(payload, ":", 3)
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return nil, fmt.Errorf("expected token to be in the format: Callback <service>:<executionRunId>:<token>")
	}

	return &CallbackAuth{
		Service:        parts[0],
		ExecutionRunID: parts[1],
		Token:          parts[2],
	}, nil
}

// IsCallbackAuth returns true if the authorization header uses the Callback scheme.
func IsCallbackAuth(authorization string) bool {
	return strings.HasPrefix(authorization, "Callback ")
}
