package authorizers

// IndeterminateError marks an error for which no authorization decision could be reached — a DB
// failure, timeout, or other unexpected lookup error — as opposed to an authoritative access
// decision (a genuine deny). The caller can propagate it as an uncached HTTP 500 instead of
// letting API Gateway cache a 403 for its TTL.
type IndeterminateError struct {
	err error
}

// NewIndeterminateError wraps err to mark it as indeterminate (not an authoritative deny).
func NewIndeterminateError(err error) *IndeterminateError {
	return &IndeterminateError{err: err}
}

func (e *IndeterminateError) Error() string {
	return e.err.Error()
}

func (e *IndeterminateError) Unwrap() error {
	return e.err
}
