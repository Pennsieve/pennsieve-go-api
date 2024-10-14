package authorizers

import (
	"context"

	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
)

type UserAuthorizer struct {
	CurrentUser *pgdbModels.User
}

func NewUserAuthorizer(currentUser *pgdbModels.User) Authorizer {
	return &UserAuthorizer{currentUser}
}

func (u *UserAuthorizer) GenerateClaims(ctx context.Context) (map[string]interface{}, error) {
	userClaim := user.Claim{
		Id:           u.CurrentUser.Id,
		NodeId:       u.CurrentUser.NodeId,
		IsSuperAdmin: u.CurrentUser.IsSuperAdmin,
	}
	return map[string]interface{}{
		"user_claim": userClaim,
	}, nil
}
