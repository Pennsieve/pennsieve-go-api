package handler

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectHandler_MissingUserNodeID(t *testing.T) {
	resp, err := DirectHandler(context.Background(), DirectAuthorizeRequest{})
	assert.NoError(t, err)
	assert.False(t, resp.IsAuthorized)
	assert.Equal(t, "user_node_id is required", resp.Error)
}

func TestDirectHandler_DatasetWithoutOrg(t *testing.T) {
	resp, err := DirectHandler(context.Background(), DirectAuthorizeRequest{
		UserNodeID:    "N:user:test",
		DatasetNodeID: "N:dataset:test",
	})
	assert.NoError(t, err)
	assert.False(t, resp.IsAuthorized)
	assert.Equal(t, "organization_node_id is required when dataset_node_id is provided", resp.Error)
}
