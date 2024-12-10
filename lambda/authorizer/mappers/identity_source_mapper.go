package mappers

import (
	"errors"
	"fmt"
	"github.com/pennsieve/pennsieve-go-api/authorizer/helpers"
)

type IdentitySource struct {
	Token string
	Other *string
}

type Mapper interface {
	// Create returns an IdentitySource or a non-nil error if one cannot be created.
	// The returned IdentitySource.Token will be a non-emtpy Token including the initial 'Bearer'.
	// If the returned IdentitySource.Other is non-nil, then it is also non-empty.
	Create() (IdentitySource, error)
}

type IdentitySourceMapper struct {
	IdentitySource []string
}

func NewIdentitySourceMapper(identitySource []string) Mapper {
	return &IdentitySourceMapper{IdentitySource: identitySource}
}

func (i *IdentitySourceMapper) Create() (IdentitySource, error) {
	m := IdentitySource{}

	if idSourceLen := len(i.IdentitySource); idSourceLen == 0 {
		return m, errors.New("identity source emtpy")
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
