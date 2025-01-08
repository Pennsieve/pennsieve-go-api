package authorizers_test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test"
	"github.com/pennsieve/pennsieve-go-api/authorizer/test/mocks"
	coreAuthorizer "github.com/pennsieve/pennsieve-go-core/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dydb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	"github.com/stretchr/testify/require"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/stretchr/testify/assert"
)

func TestManifestAuthorizer(t *testing.T) {
	manifestId := uuid.NewString()
	datasetId := int64(999)
	datasetNodeId := fmt.Sprintf("N:dataset:%s", uuid.NewString())
	authorizer := authorizers.NewManifestAuthorizer(manifestId)
	mockPg := mocks.NewMockPennsievePgAPI()
	mockDy := mocks.NewMockPennsieveDyAPI()
	token := test.NewJWTBuilder().Build(t)
	tokenClientId := uuid.NewString()
	manifestTableName := uuid.NewString()
	claimsManager := manager.NewClaimsManager(mockPg, mockDy, token.Token, tokenClientId, manifestTableName)

	currentUser := test.NewUser(101, 1001)
	orgClaim := &organization.Claim{
		Role:            pgdb.Read,
		IntId:           currentUser.PreferredOrg,
		NodeId:          fmt.Sprintf("N:organization:%s", uuid.NewString()),
		EnabledFeatures: nil,
	}
	manifest := &dydb.ManifestTable{
		ManifestId:     manifestId,
		DatasetId:      datasetId,
		DatasetNodeId:  datasetNodeId,
		OrganizationId: currentUser.PreferredOrg,
		UserId:         currentUser.Id,
		Status:         "",
		DateCreated:    0,
	}
	datasetClaim := &dataset.Claim{
		Role:   role.Viewer,
		NodeId: datasetNodeId,
		IntId:  datasetId,
	}
	mockPg.OnGetByCognitoId(token.Username).Return(currentUser, nil)
	mockPg.OnGetOrganizationClaim(currentUser.Id, currentUser.PreferredOrg).Return(orgClaim, nil)
	mockDy.OnGetManifestById(manifestTableName, manifestId).Return(manifest, nil)
	mockPg.OnGetDatasetClaim(currentUser, datasetNodeId, currentUser.PreferredOrg).Return(datasetClaim, nil)

	claims, err := authorizer.GenerateClaims(context.Background(), claimsManager, "")
	require.NoError(t, err)

	mockPg.AssertExpectations(t)
	mockDy.AssertExpectations(t)

	assert.Equal(t, 3, len(claims))
	assert.Equal(t,
		expectedUserClaim(currentUser),
		claims[coreAuthorizer.LabelUserClaim])
	assert.Equal(t,
		orgClaim,
		claims[coreAuthorizer.LabelOrganizationClaim])
	assert.Equal(t,
		datasetClaim,
		claims[coreAuthorizer.LabelDatasetClaim])
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

func expectedUserClaim(currentUser *pgdb.User) user.Claim {
	return user.Claim{
		Id:           currentUser.Id,
		NodeId:       currentUser.NodeId,
		IsSuperAdmin: currentUser.IsSuperAdmin,
	}
}
