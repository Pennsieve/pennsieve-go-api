package authorizers

import (
	"context"
	"fmt"
	coreAuthorizer "github.com/pennsieve/pennsieve-go-core/pkg/authorizer"

	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
)

type WorkspaceAuthorizer struct {
	WorkspaceID string
}

func NewWorkspaceAuthorizer(workspaceID string) Authorizer {
	return &WorkspaceAuthorizer{workspaceID}
}

func (w *WorkspaceAuthorizer) GenerateClaims(ctx context.Context, claimsManager manager.IdentityManager, authorizerMode string) (map[string]interface{}, error) {
	// Get current user
	currentUser, err := claimsManager.GetCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get current user: %w", err)
	}

	// Get Active Org
	orgInt := claimsManager.GetActiveOrg(ctx, currentUser)

	// Get Workspace Claim
	orgClaim, err := claimsManager.GetOrgClaim(ctx, currentUser, orgInt)
	if err != nil {
		return nil, fmt.Errorf("unable to get Organization Role: %w", err)
	}

	// Get Publisher's Claim
	teamClaims, err := claimsManager.GetTeamClaims(ctx, currentUser)
	if err != nil {
		return nil, fmt.Errorf("unable to get Team Claims for user: %d organization: %d: %w",
			currentUser.Id, orgInt, err)
	}

	// Get User Claim
	userClaim := claimsManager.GetUserClaim(ctx, currentUser)

	return map[string]interface{}{
		coreAuthorizer.LabelUserClaim:         userClaim,
		coreAuthorizer.LabelOrganizationClaim: orgClaim,
		coreAuthorizer.LabelTeamClaims:        teamClaims,
	}, nil
}
