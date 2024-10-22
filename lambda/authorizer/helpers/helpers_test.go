package helpers_test

import (
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/helpers"
	"github.com/stretchr/testify/assert"
)

func TestHelpers(t *testing.T) {
	UserIdentitySource := []string{"Bearer eyJra.some.random.string"}
	result := helpers.Matches(UserIdentitySource[0], `Bearer (?P<token>.*)`)
	assert.Equal(t, result, true)

	UserIdentitySource2 := []string{"earer eyJra.some.random.string"}
	result = helpers.Matches(UserIdentitySource2[0], `Bearer (?P<token>.*)`)
	assert.Equal(t, result, false)
}
