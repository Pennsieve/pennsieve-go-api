package mappers_test

import (
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/mappers"
	"github.com/stretchr/testify/assert"
)

func TestIdentitySourceMapper(t *testing.T) {
	token := "Bearer eyJra.some.random.string"
	datasetId := "N:dataset:some-uuid"

	userIdentitySource := []string{token}
	datasetIdentitySource := []string{token, datasetId}
	datasetIdentitySourceFlippedOrder := []string{datasetId, token}

	// happy path tests
	for scenario, params := range map[string]struct {
		idSource []string
		expected mappers.IdentitySource
	}{
		"token only identity source": {userIdentitySource, mappers.IdentitySource{
			Token: token,
		}},
		"identity source with additional source": {datasetIdentitySource, mappers.IdentitySource{
			Token: token,
			Other: &datasetId,
		}},
		"identity source with additional source in flipped order": {datasetIdentitySourceFlippedOrder, mappers.IdentitySource{
			Token: token,
			Other: &datasetId,
		}},
	} {
		t.Run(scenario, func(t *testing.T) {
			identitySourceMapper, err := mappers.NewIdentitySourceMapper(params.idSource)
			require.NoError(t, err)
			auxiliaryIdentitySource, err := identitySourceMapper.Create()
			require.NoError(t, err)
			assert.Equal(t, params.expected, auxiliaryIdentitySource)
		})
	}

	userTokenMissingBearer := []string{"eyJra.some.random.string"}
	userTokenMissingBearerWithOtherId := []string{"eyJra.some.random.string", datasetId}
	userTokenMissingToken := []string{"Bearer"}
	userTokenMissingTokenWithOtherId := []string{"Bearer", datasetId}
	otherIdEmpty := []string{token, ""}

	// error tests
	for scenario, params := range map[string]struct {
		idSource          []string
		expectedErrorText string
	}{
		"user token missing 'Bearer'":                  {userTokenMissingBearer, "no valid user token found"},
		"user token missing 'Bearer' with other param": {userTokenMissingBearerWithOtherId, "no valid user token found"},
		"user token missing token":                     {userTokenMissingToken, "no valid user token found"},
		"user token missing token with other param":    {userTokenMissingTokenWithOtherId, "no valid user token found"},
		"empty non-token source":                       {otherIdEmpty, "invalid non-token identity source found"},
	} {
		t.Run(scenario, func(t *testing.T) {
			identitySourceMapper, err := mappers.NewIdentitySourceMapper(params.idSource)
			require.NoError(t, err)
			_, err = identitySourceMapper.Create()
			assert.ErrorContains(t, err, params.expectedErrorText)
		})

	}

}
