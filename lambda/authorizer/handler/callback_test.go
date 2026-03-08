package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetValidatorArn(t *testing.T) {
	t.Setenv("CALLBACK_VALIDATOR_WORKFLOW_SERVICE", "arn:aws:lambda:us-east-1:123:function:workflow-validator")

	arn, err := getValidatorArn("workflow-service")
	assert.NoError(t, err)
	assert.Equal(t, "arn:aws:lambda:us-east-1:123:function:workflow-validator", arn)
}

func TestGetValidatorArnUnknownService(t *testing.T) {
	_, err := getValidatorArn("unknown-service")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no callback validator configured for service: unknown-service")
}