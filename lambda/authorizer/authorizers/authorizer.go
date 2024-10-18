package authorizers

import (
	"context"

	"github.com/pennsieve/pennsieve-go-api/authorizer/manager"
)

type Authorizer interface {
	GenerateClaims(context.Context, manager.IdentityManager, string) (map[string]interface{}, error)
}
