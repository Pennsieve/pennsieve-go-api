package dbTable

import (
	"database/sql"
	"fmt"
	"time"
)

type DbPermission int64

const (
	NoPermission DbPermission = 0
	Collaborate               = 1
	Read                      = 2
	Write                     = 4
	Delete                    = 8
	Administer                = 16
	Owner                     = 32
)

type OrganizationUser struct {
	OrganizationId int64        `json:"organization_id"`
	UserId         int64        `json:"user_id"`
	DbPermission   DbPermission `json:"permission_bit"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
}

func (o *OrganizationUser) GetByUserId(db *sql.DB, id int64) (*OrganizationUser, error) {

	queryStr := "SELECT organization_id, user_id, permission_bit, created_at, updated_at " +
		"FROM pennsieve.organization_user WHERE user_id=$1;"

	var orgUser OrganizationUser
	row := db.QueryRow(queryStr, id)
	err := row.Scan(
		&orgUser.OrganizationId,
		&orgUser.UserId,
		&orgUser.DbPermission,
		&orgUser.CreatedAt,
		&orgUser.UpdatedAt)

	switch err {
	case sql.ErrNoRows:
		fmt.Println("No rows were returned!")
		return nil, err
	case nil:
		return &orgUser, nil
	default:
		panic(err)
	}
}
