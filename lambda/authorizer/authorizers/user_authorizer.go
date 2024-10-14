package authorizers

type UserAuthorizer struct {
}

func NewUserAuthorizer() Authorizer {
	return &UserAuthorizer{}
}

func (d *UserAuthorizer) GenerateClaims() map[string]interface{} {
	return map[string]interface{}{
		"user_claim": nil,
	}
}
