package authorizers_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/pennsieve/pennsieve-go-api/authorizer/authorizers"
	"github.com/pennsieve/pennsieve-go-api/authorizer/mocks"
	"github.com/stretchr/testify/assert"
)

func TestUserAuthorizer(t *testing.T) {
	authorizer := authorizers.NewUserAuthorizer()
	claimsManager := mocks.NewMockClaimManager()
	claims, _ := authorizer.GenerateClaims(context.Background(), claimsManager)

	assert.Equal(t, len(claims), 1)
	assert.Equal(t, fmt.Sprintf("%s", claims["user_claim"]),
		"User: 1 - N:user:someRandomUuid | isSuperAdmin: true")
}
