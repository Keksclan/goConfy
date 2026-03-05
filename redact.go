package goconfy

import (
	"encoding/json"
	"fmt"

	"github.com/keksclan/goConfy/internal/redact"
)

// Redacted returns a redacted copy of the configuration struct as a generic value (usually a map).
// It recursively scans the struct and replaces sensitive fields with the string "[REDACTED]".
//
// Field selection for redaction is based on:
//   - Struct tags: Fields with `secret:"true"` are always redacted.
//   - Explicit paths: Paths provided via [WithRedactPaths] are redacted.
//   - Convention (optional): Fields with names like "password", "token", etc., can be redacted (see [WithRedactByConventionOption]).
//
// The result is safe to be printed or serialized without exposing secrets.
func Redacted[T any](cfg T, opts ...RedactOption) any {
	rc := &redactConfig{}
	for _, opt := range opts {
		opt(rc)
	}
	return redact.Apply(cfg, redact.Options{
		Paths:        rc.paths,
		ByConvention: rc.byConvention,
	})
}

// DumpRedactedJSON returns a JSON string of the redacted config.
func DumpRedactedJSON[T any](cfg T, opts ...RedactOption) (string, error) {
	redacted := Redacted(cfg, opts...)
	data, err := json.MarshalIndent(redacted, "", "  ")
	if err != nil {
		return "", fmt.Errorf("goconfy: marshal redacted JSON: %w", err)
	}
	return string(data), nil
}
