package authorizers

import (
	"context"
	"errors"
	"fmt"
	coreAuthorizer "github.com/pennsieve/pennsieve-go-core/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"

	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
)

// will be deprecated
type ManifestAuthorizer struct {
	ManifestID string
}

func NewManifestAuthorizer(manifestId string) Authorizer {
	return &ManifestAuthorizer{manifestId}
}

func (m *ManifestAuthorizer) GenerateClaims(ctx context.Context, claimsManager manager.IdentityManager, authorizerMode string) (map[string]interface{}, error) {
	// Get current user
	currentUser, err := claimsManager.GetCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get current user: %w", err)
	}

	// Get Manifest
	manifest, err := claimsManager.GetManifest(ctx, m.ManifestID)
	if err != nil {
		return nil, fmt.Errorf("error getting manifest %s: %w", m.ManifestID, err)
	}
	manifestOrgId := manifest.OrganizationId
	if tokenWorkspace, hasTokenWorkspace := claimsManager.GetTokenWorkspace(); hasTokenWorkspace && tokenWorkspace.Id != manifestOrgId {
		return nil, fmt.Errorf("manifest workspace id %d does not match API token workspace id %d",
			manifestOrgId,
			tokenWorkspace.Id)
	}
	datasetID := manifest.DatasetNodeId

	// Get Workspace Claim
	orgClaim, err := claimsManager.GetOrgClaim(ctx, currentUser.Id, manifestOrgId)
	if err != nil {
		return nil, fmt.Errorf("unable to get Organization Role: %w", err)
	}
	if orgClaim.Role == pgdb.NoPermission {
		return nil, errors.New("user has no access to workspace")
	}

	// Get Dataset Claim
	datasetClaim, err := claimsManager.GetDatasetClaim(ctx, currentUser, datasetID, manifestOrgId)
	if err != nil {
		return nil, fmt.Errorf("unable to get Dataset Role: %w", err)
	}
	// If user has no role on provided dataset --> return
	if datasetClaim.Role == role.None {
		return nil, errors.New("user has no access to dataset")
	}

	// Get User Claim
	userClaim := claimsManager.GetUserClaim(ctx, currentUser)

	if authorizerMode == "LEGACY" {
		// Get Publisher's Claim
		teamClaims, err := claimsManager.GetTeamClaims(ctx, currentUser.Id)
		if err != nil {
			return nil, fmt.Errorf("unable to get Team Claims for user: %d organization: %d: %w",
				currentUser.Id, manifestOrgId, err)
		}

		return map[string]interface{}{
			coreAuthorizer.LabelUserClaim:         userClaim,
			coreAuthorizer.LabelOrganizationClaim: orgClaim,
			coreAuthorizer.LabelDatasetClaim:      datasetClaim,
			coreAuthorizer.LabelTeamClaims:        teamClaims,
		}, nil
	}

	return map[string]interface{}{
		coreAuthorizer.LabelUserClaim:         userClaim,
		coreAuthorizer.LabelOrganizationClaim: orgClaim,
		coreAuthorizer.LabelDatasetClaim:      datasetClaim,
	}, nil
}
