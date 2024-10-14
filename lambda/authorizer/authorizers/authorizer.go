package authorizers

type Authorizer interface {
	GenerateClaims() map[string]interface{}
}
