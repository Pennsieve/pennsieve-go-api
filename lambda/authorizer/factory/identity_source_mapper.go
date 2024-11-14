package factory

import "github.com/pennsieve/pennsieve-go-api/authorizer/helpers"

type AuxiliaryIdentitySource struct {
	IdentitySource []string
	HasManisfestId bool
}

type Mapper interface {
	Create() map[string]string
}

func NewIdentitySourceMapper(identitySource []string, hasManifestId bool) Mapper {
	return &AuxiliaryIdentitySource{IdentitySource: identitySource, HasManisfestId: hasManifestId}
}

func (i *AuxiliaryIdentitySource) Create() map[string]string {
	m := make(map[string]string)
	for _, source := range i.IdentitySource {
		if helpers.Matches(source, `Bearer (?P<token>.*)`) {
			m["token"] = source
		}
		if helpers.Matches(source, `N:dataset:`) {
			m["dataset_id"] = source
		}
		if helpers.Matches(source, `N:organization:`) {
			m["workspace_id"] = source
		}
		if isManifestSource(i.HasManisfestId, source) {
			m["manifest_id"] = source
		}
	}
	return m
}

func isManifestSource(hasManifestId bool, source string) bool {
	return hasManifestId && !helpers.Matches(source, `Bearer (?P<token>.*)`) &&
		!helpers.Matches(source, `N:dataset:`) && !helpers.Matches(source, `N:organization:`)
}
