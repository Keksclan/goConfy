package goconfy

import (
	"fmt"
	"strings"
)

// FieldError represents a validation error for a specific field.
type FieldError struct {
	Field   string
	Message string

	// Optional details
	Line   int
	Column int
	Layer  string
	Path   string
}

// Error implements the error interface.
func (e *FieldError) Error() string {
	var sb strings.Builder
	if e.Layer != "" {
		sb.WriteString(fmt.Sprintf("[%s] ", e.Layer))
	}
	if e.Path != "" {
		sb.WriteString(fmt.Sprintf("path %q: ", e.Path))
	} else if e.Field != "" {
		sb.WriteString(fmt.Sprintf("field %q: ", e.Field))
	}

	if e.Line > 0 {
		sb.WriteString(fmt.Sprintf("line %d, col %d: ", e.Line, e.Column))
	}

	sb.WriteString(e.Message)
	return sb.String()
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
