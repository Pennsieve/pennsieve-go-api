package handler

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

const testJWT = "eyJraWQiOiJhYmMiLCJhbGciOiJSUzI1NiJ9.eyJzdWIiOiJ1c2VyIn0.c2lnbmF0dXJl"

func TestRedactHeaders_BearerToken(t *testing.T) {
	out := redactHeaders(map[string]string{
		"authorization": "Bearer " + testJWT,
		"user-agent":    "curl/8.4.0",
	})
	assert.Equal(t, "Bearer [redacted]", out["authorization"])
	assert.Equal(t, "curl/8.4.0", out["user-agent"])
}

func TestRedactHeaders_OtherCredentialHeaders(t *testing.T) {
	out := redactHeaders(map[string]string{
		"cookie":    "session=secret",
		"x-api-key": "abc123",
	})
	assert.Equal(t, "[redacted]", out["cookie"])
	assert.Equal(t, "[redacted]", out["x-api-key"])
}

func TestRedactHeaders_KeepsSchemeWhenTokenMissing(t *testing.T) {
	// An empty token is not a secret, and seeing it is what makes
	// "expected token to be in the format: Bearer <token>" diagnosable.
	assert.Equal(t, "Bearer", redactHeaders(map[string]string{"authorization": "Bearer"})["authorization"])
	assert.Equal(t, "Bearer", redactHeaders(map[string]string{"authorization": "Bearer   "})["authorization"])
	assert.Equal(t, "", redactHeaders(map[string]string{"authorization": ""})["authorization"])
}

func TestRedactHeaders_TokenWithoutScheme(t *testing.T) {
	assert.Equal(t, "[redacted]", redactHeaders(map[string]string{"authorization": testJWT})["authorization"])
}

func TestRedactHeaders_CallbackScheme(t *testing.T) {
	out := redactHeaders(map[string]string{"authorization": "Callback svc:run:secret"})
	assert.Equal(t, "Callback [redacted]", out["authorization"])
}

func TestRedactHeaders_DoesNotMutateInput(t *testing.T) {
	in := map[string]string{"authorization": "Bearer " + testJWT}
	redactHeaders(in)
	assert.Equal(t, "Bearer "+testJWT, in["authorization"])
}

func TestRedactHeaders_Nil(t *testing.T) {
	assert.Nil(t, redactHeaders(nil))
}

func TestRedactIdentitySource(t *testing.T) {
	auth := "Bearer " + testJWT
	out := redactIdentitySource([]string{auth, "N:dataset:1234"}, auth)
	assert.Equal(t, []string{"Bearer [redacted]", "N:dataset:1234"}, out)
}

func TestRedactIdentitySource_Nil(t *testing.T) {
	assert.Nil(t, redactIdentitySource(nil, "Bearer x"))
}

func TestRedactToken_QueryString(t *testing.T) {
	out := redactToken(map[string]string{"token": testJWT, "dataset_id": "N:dataset:1234"})
	assert.Equal(t, "[redacted]", out["token"])
	assert.Equal(t, "N:dataset:1234", out["dataset_id"])
}
