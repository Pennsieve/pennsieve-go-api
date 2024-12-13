package authorizers_test

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test/mocks"
	coreAuthorizer "github.com/pennsieve/pennsieve-go-core/pkg/authorizer"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/stretchr/testify/assert"
)

func TestManifestAuthorizer(t *testing.T) {
	authorizer := authorizers.NewManifestAuthorizer("someDatasetId")
	claimsManager := mocks.NewMockClaimManager()
	claims, _ := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	assert.Equal(t, len(claims), 3)
	assert.Equal(t,
		mocks.MockUserClaim.String(),
		fmt.Sprintf("%s", claims[coreAuthorizer.LabelUserClaim]))
	assert.Equal(t,
		mocks.MockOrgClaim.String(),
		fmt.Sprintf("%s", claims[coreAuthorizer.LabelOrganizationClaim]))
	assert.Equal(t,
		mocks.MockDatasetClaim.String(),
		fmt.Sprintf("%s", claims[coreAuthorizer.LabelDatasetClaim]))
}

func TestManifestAuthorizerLegacy(t *testing.T) {
	authorizer := authorizers.NewManifestAuthorizer("someDatasetId")
	claimsManager := mocks.NewMockClaimManager()
	claims, _ := authorizer.GenerateClaims(context.Background(), claimsManager, "LEGACY")

	assert.Equal(t, len(claims), 4)
	assert.Equal(t,
		mocks.MockUserClaim.String(),
		fmt.Sprintf("%s", claims[coreAuthorizer.LabelUserClaim]))
	assert.Equal(t,
		mocks.MockOrgClaim.String(),
		fmt.Sprintf("%s", claims[coreAuthorizer.LabelOrganizationClaim]))
	assert.Equal(t,
		mocks.MockDatasetClaim.String(),
		fmt.Sprintf("%s", claims[coreAuthorizer.LabelDatasetClaim]))
	expectedTeamClaims := "["
	separator := ""
	for _, claim := range mocks.MockTeamClaims {
		expectedTeamClaims += fmt.Sprintf("%s%s", separator, claim)
	}
	expectedTeamClaims += "]"
	assert.Equal(t,
		expectedTeamClaims,
		fmt.Sprintf("%s", claims[coreAuthorizer.LabelTeamClaims]))
}
