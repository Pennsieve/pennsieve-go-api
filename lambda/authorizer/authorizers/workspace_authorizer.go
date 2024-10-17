package authorizers

import (
	"context"
	"fmt"

	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	log "github.com/sirupsen/logrus"
)

type WorkspaceAuthorizer struct {
	WorkspaceID string
}

func NewWorkspaceAuthorizer(workspaceID string) Authorizer {
	return &WorkspaceAuthorizer{workspaceID}
}

func (w *WorkspaceAuthorizer) GenerateClaims(ctx context.Context, claimsManager manager.IdentityManager) (map[string]interface{}, error) {
	// Get current user
	currentUser, err := claimsManager.GetCurrentUser(ctx)
	if err != nil {
		log.Error("unable to get current user")
		return nil, err
	}

	// Get Active Org
	orgInt := currentUser.PreferredOrg
	jwtOrg, hasKey := claimsManager.GetToken().Get("custom:organization_id")
	if hasKey {
		orgInt = jwtOrg.(int64)
	}

	// Get ORG Claim
	orgClaim, err := claimsManager.GetQueryHandle().GetOrganizationClaim(ctx, currentUser.Id, orgInt)
	if err != nil {
		log.Error("unable to get Organization Role")
		return nil, err
	}

	// Get Publisher's Claim
	teamClaims, err := claimsManager.GetQueryHandle().GetTeamClaims(ctx, currentUser.Id)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to get Team Claims for user: %d organization: %d",
			currentUser.Id, orgInt))
		return nil, err

	}

	userClaim := user.Claim{
		Id:           currentUser.Id,
		NodeId:       currentUser.NodeId,
		IsSuperAdmin: currentUser.IsSuperAdmin,
	}

	return map[string]interface{}{
		"user_claim":  userClaim,
		"org_claim":   orgClaim,
		"teams_claim": teamClaims,
	}, nil
}
