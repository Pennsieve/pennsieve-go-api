package authorizers

// AmbiguousError marks an error that must not be treated as a cacheable deny by the API
// Gateway authorizer response. It signals a DB failure, timeout, or other unexpected lookup
// error — as opposed to an authoritative access decision (a genuine deny) — so the caller can
// propagate it as an uncached HTTP 500 instead of letting API Gateway cache a 403 for its TTL.
type AmbiguousError struct {
	err error
}

// NewAmbiguousError wraps err to mark it as ambiguous (not an authoritative deny).
func NewAmbiguousError(err error) *AmbiguousError {
	return &AmbiguousError{err: err}
}

func (e *AmbiguousError) Error() string {
	return e.err.Error()
}

func (e *AmbiguousError) Unwrap() error {
	return e.err
}
