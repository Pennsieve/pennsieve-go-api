package authorizers

import (
	"context"

	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
)

const LabelUserClaim = "user_claim"
const LabelOrganizationClaim = "org_claim"
const LabelTeamClaims = "team_claims"
const LabelDatasetClaim = "dataset_claim"

type Authorizer interface {
	GenerateClaims(context.Context, manager.IdentityManager, string) (map[string]interface{}, error)
}
