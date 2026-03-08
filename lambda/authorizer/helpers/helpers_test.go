package helpers_test

import (
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/helpers"
	"github.com/stretchr/testify/assert"
)

func TestHelpersMatches(t *testing.T) {
	UserIdentitySource := []string{"Bearer eyJra.some.random.string"}
	result := helpers.Matches(UserIdentitySource[0], `Bearer (?P<token>.*)`)
	assert.Equal(t, result, true)

	UserIdentitySource2 := []string{"earer eyJra.some.random.string"}
	result = helpers.Matches(UserIdentitySource2[0], `Bearer (?P<token>.*)`)
	assert.Equal(t, result, false)
}

func TestGetJWT(t *testing.T) {
	result, err := helpers.GetJWT("Bearer ABCD")
	assert.Equal(t, nil, err)
	assert.Equal(t, "ABCD", string(result))

	result, err = helpers.GetJWT("Bearer")
	assert.Equal(t, "expected token to be in the format: Bearer <token>", err.Error())
	assert.Equal(t, 0, len(result))
}

func TestIsCallbackAuth(t *testing.T) {
	assert.True(t, helpers.IsCallbackAuth("Callback workflow-service:run-123:token-abc"))
	assert.False(t, helpers.IsCallbackAuth("Bearer eyJra.some.random.string"))
	assert.False(t, helpers.IsCallbackAuth(""))
}

func TestParseCallbackAuth(t *testing.T) {
	result, err := helpers.ParseCallbackAuth("Callback workflow-service:run-123:token-abc")
	assert.NoError(t, err)
	assert.Equal(t, "workflow-service", result.Service)
	assert.Equal(t, "run-123", result.ExecutionRunID)
	assert.Equal(t, "token-abc", result.Token)
}

func TestParseCallbackAuthTokenWithColons(t *testing.T) {
	result, err := helpers.ParseCallbackAuth("Callback workflow-service:run-123:token:with:colons")
	assert.NoError(t, err)
	assert.Equal(t, "workflow-service", result.Service)
	assert.Equal(t, "run-123", result.ExecutionRunID)
	assert.Equal(t, "token:with:colons", result.Token)
}

func TestParseCallbackAuthErrors(t *testing.T) {
	for name, input := range map[string]string{
		"bearer scheme":    "Bearer eyJra.some.random.string",
		"missing token":    "Callback workflow-service:run-123:",
		"missing run id":   "Callback workflow-service::token",
		"missing service":  "Callback :run-123:token",
		"only two parts":   "Callback workflow-service:run-123",
		"empty payload":    "Callback ",
		"no callback":      "something-else",
	} {
		t.Run(name, func(t *testing.T) {
			_, err := helpers.ParseCallbackAuth(input)
			assert.Error(t, err)
		})
	}
}
