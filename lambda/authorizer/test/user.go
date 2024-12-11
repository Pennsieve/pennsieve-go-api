package test

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/pennsieve/pennsieve-go-core/pkg/models/pgdb"
)

func NewUser(userId int64, preferredOrgId int64) *pgdb.User {
	return &pgdb.User{
		Id:           userId,
		NodeId:       fmt.Sprintf("N:user:%s", uuid.NewString()),
		Email:        fmt.Sprintf("%s@example.com", uuid.NewString()),
		FirstName:    uuid.NewString(),
		LastName:     uuid.NewString(),
		IsSuperAdmin: false,
		PreferredOrg: preferredOrgId,
	}
}
