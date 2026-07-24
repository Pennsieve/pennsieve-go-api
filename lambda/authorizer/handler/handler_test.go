package handler

import (
	"errors"
	"fmt"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/stretchr/testify/assert"
)

func TestIsAmbiguous_AmbiguousError(t *testing.T) {
	err := authorizers.NewAmbiguousError(errors.New("connection refused"))
	assert.True(t, isAmbiguous(err))
}

func TestIsAmbiguous_AmbiguousErrorWrappedFurther(t *testing.T) {
	// DatasetAuthorizer wraps the underlying failure before marking it ambiguous, e.g.
	// NewAmbiguousError(fmt.Errorf("unable to get Organization Role: %w", err)); the marker
	// must still be detected however many times %w wraps it afterward.
	inner := authorizers.NewAmbiguousError(fmt.Errorf("unable to resolve organization: %w", errors.New("timeout")))
	outer := fmt.Errorf("unable to get current user: %w", inner)
	assert.True(t, isAmbiguous(outer))
}

func TestIsAmbiguous_PlainError(t *testing.T) {
	assert.False(t, isAmbiguous(errors.New("user has no access to dataset")))
}
