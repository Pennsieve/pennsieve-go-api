package factory

import "github.com/pennsieve/pennsieve-go-api/authorizer/helpers"

type Mapper interface {
	Create() map[string]string
}
type IdentitySourceMapper struct {
	IdentitySource []string
	HasManifestId  bool
}

func NewIdentitySourceMapper(identitySource []string, hasManifestId bool) Mapper {
	return &IdentitySourceMapper{IdentitySource: identitySource, HasManifestId: hasManifestId}
}

func (i *IdentitySourceMapper) Create() map[string]string {
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
		if isManifestSource(i.HasManifestId, source) {
			m["manifest_id"] = source
		}
	}
	return m
}

func isManifestSource(hasManifestId bool, source string) bool {
	return hasManifestId && !helpers.Matches(source, `Bearer (?P<token>.*)`)
}
