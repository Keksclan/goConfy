package goconfy

import (
	"fmt"
	"strings"
)

// FieldError represents a validation error for a specific field.
type FieldError struct {
	Field   string
	Message string
}

// Error implements the error interface.
func (e *FieldError) Error() string {
	return fmt.Sprintf("field %q: %s", e.Field, e.Message)
}

// MultiError collects multiple errors into a single error.
type MultiError struct {
	Errors []error
}

// Error implements the error interface.
func (e *MultiError) Error() string {
	msgs := make([]string, 0, len(e.Errors))
	for _, err := range e.Errors {
		msgs = append(msgs, err.Error())
	}
	return strings.Join(msgs, "; ")
}

// Unwrap returns the list of wrapped errors.
func (e *MultiError) Unwrap() []error {
	return e.Errors
}
