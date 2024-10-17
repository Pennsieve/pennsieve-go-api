package authorizers

import (
	"context"

	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	log "github.com/sirupsen/logrus"
)

type UserAuthorizer struct{}

func NewUserAuthorizer() Authorizer {
	return &UserAuthorizer{}
}

func (u *UserAuthorizer) GenerateClaims(ctx context.Context, claimsManager manager.IdentityManager) (map[string]interface{}, error) {
	// Get current user
	currentUser, err := claimsManager.GetCurrentUser(ctx)
	if err != nil {
		log.Error("unable to get current user")
		return nil, err
	}
	userClaim := user.Claim{
		Id:           currentUser.Id,
		NodeId:       currentUser.NodeId,
		IsSuperAdmin: currentUser.IsSuperAdmin,
	}
	return map[string]interface{}{
		"user_claim": userClaim,
	}, nil
}
