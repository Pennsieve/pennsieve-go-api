package manager

import (
	"context"
	"database/sql"
	"fmt"
)

// AccessScope is every workspace a user belongs to, and every team they are on
// in those workspaces, as node IDs.
type AccessScope struct {
	WorkspaceNodeIds []string
	TeamNodeIds      []string
}

// Queryer is the part of *sql.DB that GetAccessScope uses.
type Queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

const accessScopeWorkspacesQuery = `SELECT o.node_id
	FROM pennsieve.organization_user ou
	JOIN pennsieve.organizations o ON o.id = ou.organization_id
	WHERE ou.user_id = $1 AND ou.permission_bit > 0
	  AND ($2::bigint = 0 OR o.id = $2::bigint)
	ORDER BY o.id`

const accessScopeTeamsQuery = `SELECT DISTINCT t.node_id
	FROM pennsieve.team_user tu
	JOIN pennsieve.teams t ON t.id = tu.team_id
	JOIN pennsieve.organization_team ot ON ot.team_id = t.id
	JOIN pennsieve.organization_user ou ON ou.organization_id = ot.organization_id AND ou.user_id = tu.user_id
	WHERE tu.user_id = $1 AND ou.permission_bit > 0
	  AND ($2::bigint = 0 OR ot.organization_id = $2::bigint)
	ORDER BY t.node_id`

// GetAccessScope returns the user's workspaces and teams. Unlike the claims,
// it is not limited to one workspace, because an app shared with any of them
// is visible to the user. A non-zero organizationId (an API token's workspace)
// limits it to that workspace.
func GetAccessScope(ctx context.Context, db Queryer, userId int64, organizationId int64) (AccessScope, error) {
	workspaces, err := queryNodeIds(ctx, db, accessScopeWorkspacesQuery, userId, organizationId)
	if err != nil {
		return AccessScope{}, fmt.Errorf("querying workspaces of user %d: %w", userId, err)
	}
	teams, err := queryNodeIds(ctx, db, accessScopeTeamsQuery, userId, organizationId)
	if err != nil {
		return AccessScope{}, fmt.Errorf("querying teams of user %d: %w", userId, err)
	}
	return AccessScope{WorkspaceNodeIds: workspaces, TeamNodeIds: teams}, nil
}

func queryNodeIds(ctx context.Context, db Queryer, query string, args ...any) ([]string, error) {
	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	ids := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
