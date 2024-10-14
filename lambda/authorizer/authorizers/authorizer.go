package authorizers

import "context"

type Authorizer interface {
	GenerateClaims(context.Context) (map[string]interface{}, error)
}
