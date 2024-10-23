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

func TestHelpersDecodeIdentitySource(t *testing.T) {
	datasetIdentitySource := []string{"Bearer eyJra.some.random.string", "N%3Adataset%3A3c10091e-4ef8-45ac-b3ae-4497eb34c7dc"}
	result, _ := helpers.DecodeIdentitySource(datasetIdentitySource[1])
	assert.Equal(t, result, "N:dataset:3c10091e-4ef8-45ac-b3ae-4497eb34c7dc")

	// non-encoded string should be unaffected
	datasetIdentitySource2 := []string{"Bearer eyJra.some.random.string", "N:dataset:3c10091e-4ef8-45ac-b3ae-4497eb34c7dc"}
	result, _ = helpers.DecodeIdentitySource(datasetIdentitySource2[1])
	assert.Equal(t, result, "N:dataset:3c10091e-4ef8-45ac-b3ae-4497eb34c7dc")
}
