package mappers

import (
	"errors"
	"fmt"
	"github.com/pennsieve/pennsieve-go-api/authorizer/helpers"
)

type MappedIdentitySource struct {
	Token string
	Other *string
}

type Mapper interface {
	// Create returns an MappedIdentitySource or a non-nil error if one cannot be created.
	// The returned MappedIdentitySource.Token will be a non-empty Token including the initial 'Bearer'.
	// If the returned MappedIdentitySource.Other is non-nil, then it is also non-empty.
	Create() (MappedIdentitySource, error)
}

type IdentitySourceMapper struct {
	IdentitySource []string
}

func NewIdentitySourceMapper(identitySource []string) Mapper {
	return &IdentitySourceMapper{IdentitySource: identitySource}
}

func (i *IdentitySourceMapper) Create() (MappedIdentitySource, error) {
	m := MappedIdentitySource{}

	if idSourceLen := len(i.IdentitySource); idSourceLen == 0 {
		return m, errors.New("identity source empty")
	} else if idSourceLen > 2 {
		return m, fmt.Errorf("identity source too long: %d", idSourceLen)
	}

	for _, source := range i.IdentitySource {
		if helpers.Matches(source, `Bearer (?P<token>.*)`) {
			m.Token = source
		} else {
			m.Other = &source
		}
	}
	if len(m.Token) == 0 {
		return m, fmt.Errorf("no valid user token found in %s", i.IdentitySource)
	}
	if m.Other != nil && len(*m.Other) == 0 {
		return m, fmt.Errorf("invalid non-token identity source found in %s", i.IdentitySource)
	}
	return m, nil
}
