package beautify

// ValidationError is a user-facing validation failure with a stable error code.
// Codes are safe to expose to API consumers; messages never contain internals.
type ValidationError struct {
	Code    string
	Message string
}

func (e *ValidationError) Error() string { return e.Message }

func newValidationError(code, message string) *ValidationError {
	return &ValidationError{Code: code, Message: message}
}
