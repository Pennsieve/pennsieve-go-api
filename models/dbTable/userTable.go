package dbTable

import (
	"database/sql"
	"fmt"
)

type User struct {
	Id           int64  `json:"id"`
	Email        string `json:"email"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	IsSuperAdmin bool   `json:"is_super_admin"`
	PreferredOrg int64  `json:"preferred_org_id"`
}

//GetByCognitoId returns a user from Postgress based on his/her cognito-id
//This function also returns the preferred org and whether the user is a super-admin.
func (u *User) GetByCognitoId(db *sql.DB, id string) (*User, error) {

	queryStr := "SELECT id, email, first_name, last_name, is_super_admin, preferred_org_id " +
		"FROM pennsieve.users WHERE cognito_id=$1;"

	var user User
	row := db.QueryRow(queryStr, id)
	err := row.Scan(
		&user.Id,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.IsSuperAdmin,
		&user.PreferredOrg)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return nil, err
	case nil:
		return &user, nil
	default:
		panic(err)
	}
}
