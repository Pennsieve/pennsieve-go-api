package authorizers

import (
	"context"
	"errors"
	"fmt"

	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
	coreAuthorizer "github.com/pennsieve/pennsieve-go-core/pkg/authorizer"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/role"
	corePgdb "github.com/pennsieve/pennsieve-go-core/pkg/queries/pgdb"
)

type DatasetAuthorizer struct {
	DatasetId string
}

func NewDatasetAuthorizer(datasetId string) Authorizer {
	return &DatasetAuthorizer{datasetId}
}

func (d *DatasetAuthorizer) GenerateClaims(ctx context.Context, claimsManager manager.IdentityManager, authorizerMode string) (map[string]interface{}, error) {
	// Get current user
	currentUser, err := claimsManager.GetCurrentUser(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to get current user: %w", err)
	}

	// Always resolve the dataset's org from the request via the dataset_organization map,
	// never from the user's preferred/active org.
	orgInt, err := claimsManager.GetOrganizationIdForDataset(ctx, d.DatasetId)
	if err != nil {
		var notFound corePgdb.DatasetOrganizationNotFoundError
		if errors.As(err, &notFound) {
			// Genuine map miss: the dataset doesn't exist. Clean, cacheable deny.
			return nil, fmt.Errorf("no organization found for dataset %s: %w", d.DatasetId, err)
		}
		// DB/connection failure resolving the map: indeterminate, must not be cached as a deny.
		return nil, NewIndeterminateError(fmt.Errorf("unable to resolve organization for dataset %s: %w", d.DatasetId, err))
	}

	// Token-pool (API-key) tokens are already scoped to a single org. An API key scoped to one
	// org must not reach a dataset that lives in a different org, even if the underlying user
	// has access to that dataset in its actual org.
	if tokenWorkspace, hasTokenWorkspace := claimsManager.GetTokenWorkspace(); hasTokenWorkspace {
		if tokenWorkspace.Id != orgInt {
			return nil, fmt.Errorf("token workspace %d does not match organization %d for dataset %s",
				tokenWorkspace.Id, orgInt, d.DatasetId)
		}
	}

	// Get Workspace Claim
	orgClaim, err := claimsManager.GetOrgClaim(ctx, currentUser.Id, orgInt)
	if err != nil {
		var notOrgMember corePgdb.OrganizationUserNotFoundError
		if errors.As(err, &notOrgMember) {
			// The user is not a member of this org: a clean, authoritative (cacheable) deny,
			// not a DB failure. GetOrganizationClaim's inner join returns this error rather
			// than a NoPermission claim when there's no organization_user row for the user.
			return nil, fmt.Errorf("user has no access to organization %d: %w", orgInt, err)
		}
		return nil, NewIndeterminateError(fmt.Errorf("unable to get Organization Role: %w", err))
	}

	// Get Dataset Claim
	datasetClaim, err := claimsManager.GetDatasetClaim(ctx, currentUser, d.DatasetId, orgInt)
	if err != nil {
		return nil, NewIndeterminateError(fmt.Errorf("unable to get Dataset Role: %w", err))
	}
	// If user has no role on provided dataset --> return
	if datasetClaim.Role == role.None {
		return nil, errors.New("user has no access to dataset")
	}

	// Get User Claim
	userClaim := claimsManager.GetUserClaim(ctx, currentUser)

	if authorizerMode == "LEGACY" {
		// Get Publisher's Claim
		teamClaims, err := claimsManager.GetTeamClaimsForOrg(ctx, currentUser.Id, orgInt)
		if err != nil {
			return nil, NewIndeterminateError(fmt.Errorf("unable to get Team Claims for user: %d organization: %d: %w",
				currentUser.Id, orgInt, err))
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
