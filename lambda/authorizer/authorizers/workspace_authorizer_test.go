package authorizers_test

import (
	"context"
	"fmt"
	coreAuthorizer "github.com/pennsieve/pennsieve-go-core/pkg/authorizer"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/pennsieve/pennsieve-go-api/authorizer/mocks"
	"github.com/stretchr/testify/assert"
)

func TestWorkspaceAuthorizer(t *testing.T) {
	authorizer := authorizers.NewWorkspaceAuthorizer("someWorkspaceId")
	claimsManager := mocks.NewMockClaimManager()
	claims, _ := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	assert.Equal(t, len(claims), 3)
	assert.Equal(t,
		mocks.MockUserClaim.String(),
		fmt.Sprintf("%s", claims[coreAuthorizer.LabelUserClaim]))
	assert.Equal(t,
		mocks.MockOrgClaim.String(),
		fmt.Sprintf("%s", claims[coreAuthorizer.LabelOrganizationClaim]))
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
