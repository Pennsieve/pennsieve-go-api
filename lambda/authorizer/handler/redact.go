package handler

import "strings"

const redacted = "[redacted]"

// sensitiveHeaders are request headers whose values are credentials. API Gateway
// lower-cases header names in the authorizer event.
var sensitiveHeaders = map[string]bool{
	"authorization": true,
	"cookie":        true,
	"x-api-key":     true,
}

// redactCredential hides the secret part of a credential header value but keeps
// the scheme ("Bearer", "Callback"), so a malformed or empty header is still
// diagnosable from the logs.
func redactCredential(v string) string {
	scheme, rest, found := strings.Cut(v, " ")
	if !found {
		// A bare scheme such as "Bearer" carries no secret; anything else might.
		if strings.EqualFold(v, "bearer") || strings.EqualFold(v, "callback") || v == "" {
			return v
		}
		return redacted
	}
	if strings.TrimSpace(rest) == "" {
		return scheme
	}
	return scheme + " " + redacted
}

// redactHeaders returns a copy of headers with credential values redacted.
func redactHeaders(headers map[string]string) map[string]string {
	if headers == nil {
		return nil
	}
	out := make(map[string]string, len(headers))
	for k, v := range headers {
		if sensitiveHeaders[strings.ToLower(k)] {
			out[k] = redactCredential(v)
			continue
		}
		out[k] = v
	}
	return out
}

// redactIdentitySource returns a copy of the identity source with any entry
// holding the Authorization header value redacted. Other entries (dataset_id,
// organization_id, manifest_id) are node ids and are kept.
func redactIdentitySource(identitySource []string, authorization string) []string {
	if identitySource == nil {
		return nil
	}
	out := make([]string, len(identitySource))
	for i, v := range identitySource {
		if authorization != "" && v == authorization {
			out[i] = redactCredential(v)
			continue
		}
		out[i] = v
	}
	return out
}
