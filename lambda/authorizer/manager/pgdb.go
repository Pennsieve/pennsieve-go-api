package manager

import (
	"context"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/dataset"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/organization"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/teamUser"
)

// PennsievePgAPI is an interface only containing the methods of *pgdb.Queries that are used by the ClaimsManager.
type PennsievePgAPI interface {
	GetDatasetClaim(ctx context.Context, user *pgdb.User, datasetNodeId string, organizationId int64) (*dataset.Claim, error)
	GetOrganizationClaim(ctx context.Context, userId int64, organizationId int64) (*organization.Claim, error)
	GetOrganizationClaimByNodeId(ctx context.Context, userId int64, organizationNodeId string) (*organization.Claim, error)
	GetTeamClaims(ctx context.Context, userId int64) ([]teamUser.Claim, error)
	// GetUserByCognitoId returns a Pennsieve User based on the cognito id in the token pool.
	GetUserByCognitoId(ctx context.Context, cognitoId string) (*pgdb.User, error)
	// GetByCognitoId returns a Pennsieve User based on the cognito id in the users table.
	GetByCognitoId(ctx context.Context, cognitoId string) (*pgdb.User, error)
}
