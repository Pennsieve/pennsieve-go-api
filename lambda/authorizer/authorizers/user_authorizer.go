package authorizers

import (
	"context"
	"fmt"

	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
)

type UserAuthorizer struct{}

func NewUserAuthorizer() Authorizer {
	return &UserAuthorizer{}
}

func (u *UserAuthorizer) GenerateClaims(ctx context.Context, claimsManager manager.IdentityManager, authorizerMode string) (map[string]interface{}, error) {
	// Get current user
	currentUser, err := claimsManager.GetCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get current user: %w", err)
	}
	// Get User Claim
	userClaim := claimsManager.GetUserClaim(ctx, currentUser)

	if authorizerMode == "LEGACY" {
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

		return map[string]interface{}{
			"user_claim":  userClaim,
			"org_claim":   orgClaim,
			"teams_claim": teamClaims,
		}, nil
	}

	return map[string]interface{}{
		"user_claim": userClaim,
	}, nil
}
