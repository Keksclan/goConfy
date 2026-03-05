package goconfy

import (
	"fmt"
	"strings"
)

// FieldError represents an error that occurred during a specific phase of the configuration pipeline.
// It provides contextual information like the field path, line and column numbers,
// and the pipeline layer where the error occurred.
type FieldError struct {
	// Field is the name of the struct field associated with the error (if applicable).
	Field string
	// Message is the human-readable error message.
	Message string

	// Line is the 1-based line number in the YAML source.
	Line int
	// Column is the 1-based column number in the YAML source.
	Column int
	// Layer is the pipeline phase where the error occurred (e.g., "base", "dotenv", "profile", "validation").
	Layer string
	// Path is the dot-separated path to the configuration key (e.g., "db.password").
	Path string
}

// Error implements the error interface, providing a formatted string with all available context.
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
		if e.Column > 0 {
			sb.WriteString(fmt.Sprintf("line %d, col %d: ", e.Line, e.Column))
		} else {
			sb.WriteString(fmt.Sprintf("line %d: ", e.Line))
		}
	}

	sb.WriteString(e.Message)
	return sb.String()
}

// MultiError collects multiple errors into a single error.
// It implements errors.Join compatibility by providing an Unwrap() []error method.
type MultiError struct {
	Errors []error
}

// Error implements the error interface.
func (e *MultiError) Error() string {
	if len(e.Errors) == 0 {
		return ""
	}
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
