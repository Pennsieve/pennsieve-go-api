package authorizers_test

import (
	"testing"
)

func TestDatasetAuthorizer(t *testing.T) {
	// currentUser := &pgdbModels.User{Id: 1, NodeId: "N:user:someRandomUuid", IsSuperAdmin: true}
	// DatasetIdentitySource := []string{"Bearer eyJra.some.random.string", "N:dataset:some-uuid"}
	// token := jwt.New()
	// authorizer := authorizers.NewDatasetAuthorizer(currentUser, nil, DatasetIdentitySource, token)
	// claims, _ := authorizer.GenerateClaims(context.Background())
	// log.Println(claims["user_claim"])

	// assert.Equal(t, len(claims), 1)
	// assert.Equal(t, fmt.Sprintf("%s", claims["user_claim"]),
	// 	"User: 1 - N:user:someRandomUuid | isSuperAdmin: true")
}
