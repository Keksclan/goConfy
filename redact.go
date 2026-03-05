package goconfy

import (
	"encoding/json"
	"fmt"

	"github.com/keksclan/goConfy/internal/redact"
)

// Redacted returns a redacted copy of the config struct as a generic value.
// Fields tagged with `secret:"true"` and any explicitly configured paths are
// replaced with "[REDACTED]".
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
