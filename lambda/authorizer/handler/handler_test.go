package handler

import (
	"errors"
	"fmt"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/stretchr/testify/assert"
)

func TestIsIndeterminate_IndeterminateError(t *testing.T) {
	err := authorizers.NewIndeterminateError(errors.New("connection refused"))
	assert.True(t, isIndeterminate(err))
}

func TestIsIndeterminate_IndeterminateErrorWrappedFurther(t *testing.T) {
	// DatasetAuthorizer wraps the underlying failure before marking it indeterminate, e.g.
	// NewIndeterminateError(fmt.Errorf("unable to get Organization Role: %w", err)); the marker
	// must still be detected however many times %w wraps it afterward.
	inner := authorizers.NewIndeterminateError(fmt.Errorf("unable to resolve organization: %w", errors.New("timeout")))
	outer := fmt.Errorf("unable to get current user: %w", inner)
	assert.True(t, isIndeterminate(outer))
}

func TestIsIndeterminate_PlainError(t *testing.T) {
	assert.False(t, isIndeterminate(errors.New("user has no access to dataset")))
}