package authorizers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/pennsieve/pennsieve-go-api/authorizer/mocks"
	"github.com/stretchr/testify/assert"
)

func TestDatasetAuthorizer(t *testing.T) {
	authorizer := authorizers.NewDatasetAuthorizer("someDatasetId")
	claimsManager := mocks.NewMockClaimManager()
	claims, _ := authorizer.GenerateClaims(context.Background(), claimsManager, "")

	assert.Equal(t, len(claims), 3)
	assert.Equal(t, fmt.Sprintf("%s", claims[authorizers.LabelUserClaim]),
		"User: 1 - N:user:someRandomUuid | isSuperAdmin: true")
	assert.Equal(t, fmt.Sprintf("%s", claims[authorizers.LabelOrganizationClaim]),
		"OrganizationId: 0 - NoPermission")
	assert.Equal(t, fmt.Sprintf("%s", claims[authorizers.LabelDatasetClaim]),
		" (0) - Manager")
}

func TestDatasetAuthorizerLegacy(t *testing.T) {
	authorizer := authorizers.NewDatasetAuthorizer("someDatasetId")
	claimsManager := mocks.NewMockClaimManager()
	claims, _ := authorizer.GenerateClaims(context.Background(), claimsManager, "LEGACY")

	assert.Equal(t, len(claims), 4)
	assert.Equal(t, fmt.Sprintf("%s", claims[authorizers.LabelUserClaim]),
		"User: 1 - N:user:someRandomUuid | isSuperAdmin: true")
	assert.Equal(t, fmt.Sprintf("%s", claims[authorizers.LabelOrganizationClaim]),
		"OrganizationId: 0 - NoPermission")
	assert.Equal(t, fmt.Sprintf("%s", claims[authorizers.LabelDatasetClaim]),
		" (0) - Manager")
	assert.Equal(t, fmt.Sprintf("%s", claims[authorizers.LabelTeamClaims]),
		"[Name: someTeam1 (id: 1 nodeId:  permission: 0)]")
}
