package authorizers

import (
	"context"
	"fmt"

	"github.com/lestrrat-go/jwx/v2/jwt"
	pgdbModels "github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/user"
	pgdbQueries "github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
	log "github.com/sirupsen/logrus"
)

type WorkspaceAuthorizer struct {
	CurrentUser    *pgdbModels.User
	Queries        *pgdbQueries.Queries
	IdentitySource []string
	Token          jwt.Token
}

func NewWorkspaceAuthorizer(currentUser *pgdbModels.User, pddb *pgdbQueries.Queries, IdentitySource []string, token jwt.Token) Authorizer {
	return &WorkspaceAuthorizer{currentUser, pddb, IdentitySource, token}
}

func (w *WorkspaceAuthorizer) GenerateClaims(ctx context.Context) (map[string]interface{}, error) {
	// Get Active Org
	orgInt := w.CurrentUser.PreferredOrg
	jwtOrg, hasKey := w.Token.Get("custom:organization_id")
	if hasKey {
		orgInt = jwtOrg.(int64)
	}

	// Get ORG Claim
	orgClaim, err := w.Queries.GetOrganizationClaim(ctx, w.CurrentUser.Id, orgInt)
	if err != nil {
		log.Error("unable to get Organization Role")
		return nil, err
	}

	// Get Publisher's Claim
	teamClaims, err := w.Queries.GetTeamClaims(ctx, w.CurrentUser.Id)
	if err != nil {
		log.Error(fmt.Sprintf("Unable to get Team Claims for user: %d organization: %d",
			w.CurrentUser.Id, orgInt))
		return nil, err

	}

	userClaim := user.Claim{
		Id:           w.CurrentUser.Id,
		NodeId:       w.CurrentUser.NodeId,
		IsSuperAdmin: w.CurrentUser.IsSuperAdmin,
	}

	return map[string]interface{}{
		"user_claim":  userClaim,
		"org_claim":   orgClaim,
		"teams_claim": teamClaims,
	}, nil
}
