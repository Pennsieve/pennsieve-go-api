package authorizers_test

import (
	"context"
	"fmt"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test/mocks"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/stretchr/testify/assert"
)

func TestUserAuthorizer(t *testing.T) {
	authorizer := authorizers.NewUserAuthorizer()
	claimsManager := mocks.NewMockClaimManager()
	claims, _ := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	assert.Equal(t, len(claims), 1)
	assert.Equal(t, fmt.Sprintf("%s", claims[authorizers.LabelUserClaim]),
		"User: 1 - N:user:someRandomUuid | isSuperAdmin: true")
}

func TestUserAuthorizerLegacy(t *testing.T) {
	authorizer := authorizers.NewUserAuthorizer()
	claimsManager := mocks.NewMockClaimManager()
	claims, _ := authorizer.GenerateClaims(context.Background(), claimsManager, "LEGACY")

	assert.Equal(t, len(claims), 3)
	assert.Equal(t, fmt.Sprintf("%s", claims[authorizers.LabelUserClaim]),
		"User: 1 - N:user:someRandomUuid | isSuperAdmin: true")
	assert.Equal(t, fmt.Sprintf("%s", claims[authorizers.LabelOrganizationClaim]),
		"OrganizationId: 0 - NoPermission")
	assert.Equal(t, fmt.Sprintf("%s", claims[authorizers.LabelTeamClaims]),
		"[Name: someTeam1 (id: 1 nodeId:  permission: 0)]")
}
