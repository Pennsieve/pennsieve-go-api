package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectHandler_MissingUserID(t *testing.T) {
	resp, err := DirectHandler(context.Background(), DirectAuthorizeRequest{})
	assert.NoError(t, err)
	assert.False(t, resp.IsAuthorized)
	assert.Equal(t, "user_id is required", resp.Error)
}

func TestDirectHandler_DatasetWithoutOrg(t *testing.T) {
	resp, err := DirectHandler(context.Background(), DirectAuthorizeRequest{
		UserID:    1,
		DatasetID: "N:dataset:test",
	})
	assert.NoError(t, err)
	assert.False(t, resp.IsAuthorized)
	assert.Equal(t, "organization_id is required when dataset_id is provided", resp.Error)
}
