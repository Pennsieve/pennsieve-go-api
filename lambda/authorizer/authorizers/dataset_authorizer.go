package authorizers

type DatasetAuthorizer struct {
}

func NewDatasetAuthorizer() Authorizer {
	return &DatasetAuthorizer{}
}

func (d *DatasetAuthorizer) GenerateClaims() map[string]interface{} {
	return map[string]interface{}{
		"user_claim":    nil,
		"dataset_claim": nil,
	}
}
