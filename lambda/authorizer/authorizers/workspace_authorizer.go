package authorizers

import (
	"context"
	"fmt"

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

	if tokenWorkspace, hasTokenWorkspace := claimsManager.GetTokenWorkspace(); hasTokenWorkspace && tokenWorkspace.NodeId != w.WorkspaceID {
		return nil, fmt.Errorf("provided workspace id %s does not match API token workspace id %s",
			w.WorkspaceID,
			tokenWorkspace.NodeId)
	}

	// Get Workspace Claim
	orgClaim, err := claimsManager.GetOrgClaimByNodeId(ctx, currentUser, w.WorkspaceID)
	if err != nil {
		return nil, fmt.Errorf("unable to get Organization Role: %w", err)
	}

	// Get Publisher's Claim
	teamClaims, err := claimsManager.GetTeamClaims(ctx, currentUser)
	if err != nil {
		return nil, fmt.Errorf("unable to get Team Claims for user: %d organization: %s: %w",
			currentUser.Id, w.WorkspaceID, err)
	}

	// Get User Claim
	userClaim := claimsManager.GetUserClaim(ctx, currentUser)

	return map[string]interface{}{
		"user_claim":  userClaim,
		"org_claim":   orgClaim,
		"teams_claim": teamClaims,
	}, nil
}
