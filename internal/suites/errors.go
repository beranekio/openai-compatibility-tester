package suites

import "fmt"

// ValidationError reports a response that did not meet compatibility expectations.
type ValidationError struct {
	Suite   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Suite, e.Message)
}

func fail(suite, message string) error {
	return &ValidationError{Suite: suite, Message: message}
}
