package authorizer

import (
	"database/sql"
	"fmt"
	"github.com/pennsieve/pennsieve-go-api/models/dataset"
	"github.com/pennsieve/pennsieve-go-api/models/dbTable"
	"log"
	"sort"
)

// GetDatasetClaim returns the highest role that the user has for a given dataset.
// This method checks the roles of the dataset, the teams, and the specific user roles.
func GetDatasetClaim(db *sql.DB, user *dbTable.User, datasetNodeId string, organizationId int64) (*dataset.Claim, error) {

	// if user is super-admin
	if user.IsSuperAdmin {
		// USER IS A SUPER-ADMIN

		//TODO: HANDLE SPECIAL CASE
		fmt.Println("Not handling super-user authorization at this point.")

	}

	// USER IS NOT A SUPER-ADMIN

	// 1. Get Dataset Role and integer ID
	datasetQuery := fmt.Sprintf("SELECT id, role FROM \"%d\".datasets WHERE node_id='%s';", organizationId, datasetNodeId)

	var datasetId int64
	var maybeDatasetRole sql.NullString

	row := db.QueryRow(datasetQuery)
	err := row.Scan(
		&datasetId,
		&maybeDatasetRole)

	if err != nil {
		switch err {
		case sql.ErrNoRows:
			log.Println("No rows were returned!")
			return nil, err
		default:
			log.Println("Uknown Error while scanning dataset table: ", err)
			panic(err)
		}
	}

	// If maybeDatasetRole is set, include the role, otherwise use none-role
	datasetRole := dataset.None
	if maybeDatasetRole.Valid {
		var ok bool
		datasetRole, ok = dataset.RoleFromString(maybeDatasetRole.String)
		if !ok {
			log.Fatalln("Could not map Dataset Role from database string: ", maybeDatasetRole.String)
		}
	}

	// 2. Get Team Role
	teamPermission := fmt.Sprintf("\"%d\".dataset_team.role", organizationId)
	datasetTeam := fmt.Sprintf("\"%d\".dataset_team", organizationId)
	teamQueryStr := fmt.Sprintf("SELECT %s FROM pennsieve.team_user JOIN %s "+
		"ON pennsieve.team_user.team_id = %s.team_id "+
		"WHERE user_id=%d AND dataset_id=%d", teamPermission, datasetTeam, datasetTeam, user.Id, datasetId)

	// Get User Role
	userPermission := fmt.Sprintf("\"%d\".dataset_user.role", organizationId)
	datasetUser := fmt.Sprintf("\"%d\".dataset_user", organizationId)
	userQueryStr := fmt.Sprintf("SELECT %s FROM %s WHERE user_id=%d AND dataset_id=%d",
		userPermission, datasetUser, user.Id, datasetId)

	// Combine all queries in a single Union.
	fullQuery := teamQueryStr + " UNION " + userQueryStr + ";"

	rows, err := db.Query(fullQuery)
	if err != nil {
		return nil, err
	}

	roles := []dataset.Role{
		datasetRole,
	}
	for rows.Next() {
		var roleString string
		err = rows.Scan(
			&roleString)

		if err != nil {
			log.Println("ERROR: ", err)
		}

		role, ok := dataset.RoleFromString(roleString)
		if !ok {
			log.Fatalln("Could not map Dataset Role from database string.")
		}
		roles = append(roles, role)
	}

	// Sort roles by enum value --> first entry is the highest level of permission.
	sort.Slice(roles, func(i, j int) bool {
		return roles[i] > roles[j]
	})

	// return the maximum role that the user has.
	claim := dataset.Claim{
		Role:   roles[0],
		NodeId: datasetNodeId,
		IntId:  datasetId,
	}

	return &claim, nil

}
