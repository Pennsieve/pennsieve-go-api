package authorizers

import (
	"context"
	"fmt"

	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	log "github.com/sirupsen/logrus"
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
		log.Error("unable to get current user")
		return nil, err
	}

	// Get Active Org
	orgInt := claimsManager.GetActiveOrg(ctx, currentUser)

	// Get Workspace Claim
	orgClaim, err := claimsManager.GetOrgClaim(ctx, currentUser, orgInt)
	if err != nil {
		log.Error("unable to get Organization Role")
		return nil, err
	}

	// Get Publisher's Claim
	teamClaims, err := claimsManager.GetTeamClaims(ctx, currentUser)
	if err != nil {
		log.Error(fmt.Sprintf("unable to get Team Claims for user: %d organization: %d",
			currentUser.Id, orgInt))
		return nil, err
	}

	// Get User Claim
	userClaim := claimsManager.GetUserClaim(ctx, currentUser)

	return map[string]interface{}{
		"user_claim":  userClaim,
		"org_claim":   orgClaim,
		"teams_claim": teamClaims,
	}, nil
}
