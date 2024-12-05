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

func NewIdentitySourceMapper(identitySource []string) (Mapper, error) {
	if idSourceLen := len(identitySource); idSourceLen == 0 {
		return nil, errors.New("identity source emtpy")
	} else if idSourceLen > 2 {
		return nil, fmt.Errorf("identity source too long: %d", idSourceLen)
	}
	return &IdentitySourceMapper{IdentitySource: identitySource}, nil
}

func (i *IdentitySourceMapper) Create() (IdentitySource, error) {
	m := IdentitySource{}
	for _, source := range i.IdentitySource {
		// need to avoid for-loop variable gotcha since we may take the address of source below (and we are on Go version < 1.22)
		source := source
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
