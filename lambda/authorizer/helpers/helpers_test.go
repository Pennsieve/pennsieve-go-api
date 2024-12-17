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
