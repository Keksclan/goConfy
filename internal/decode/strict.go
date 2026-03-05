package decode

import (
	"bytes"
	"fmt"
	"regexp"
	"strconv"

	"gopkg.in/yaml.v3"
)

var yamlErrRegex = regexp.MustCompile(`line (\d+)(?:, col (\d+))?: field (.+) not found in type (.+)`)

// Strict decodes YAML bytes into the target struct using strict mode,
// which rejects unknown fields.
func Strict(data []byte, target any) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	dec.KnownFields(true)
	if err := dec.Decode(target); err != nil {
		if wrapped, ok := wrapError(err); ok {
			return wrapped
		}
		return fmt.Errorf("strict YAML decode: %w", err)
	}
	return nil
}

func wrapError(err error) (error, bool) {
	msg := err.Error()
	matches := yamlErrRegex.FindStringSubmatch(msg)
	if len(matches) == 5 {
		line, _ := strconv.Atoi(matches[1])
		col := 0
		if matches[2] != "" {
			col, _ = strconv.Atoi(matches[2])
		}
		field := matches[3]
		return &fieldError{
			Line:    line,
			Column:  col,
			Field:   field,
			Message: fmt.Sprintf("unknown field %q", field),
		}, true
	}
	return nil, false
}

type fieldError struct {
	Field   string
	Message string
	Line    int
	Column  int
}

func (e *fieldError) Error() string {
	if e.Line > 0 {
		return fmt.Sprintf("line %d: field %q: %s", e.Line, e.Field, e.Message)
	}
	return fmt.Sprintf("field %q: %s", e.Field, e.Message)
}

func (e *fieldError) GetField() string { return e.Field }
func (e *fieldError) GetLine() int     { return e.Line }
func (e *fieldError) GetColumn() int   { return e.Column }

// Relaxed decodes YAML bytes into the target struct without strict mode.
func Relaxed(data []byte, target any) error {
	dec := yaml.NewDecoder(bytes.NewReader(data))
	if err := dec.Decode(target); err != nil {
		return fmt.Errorf("YAML decode: %w", err)
	}
	return nil
}
