package authorizers

type WorkspaceAuthorizer struct {
}

func NewWorkspaceAuthorizer() Authorizer {
	return &WorkspaceAuthorizer{}
}

func (d *WorkspaceAuthorizer) GenerateClaims() map[string]interface{} {
	return map[string]interface{}{
		"user_claim":       nil,
		"workspace _claim": nil,
		"teams_claim":      nil,
	}
}
