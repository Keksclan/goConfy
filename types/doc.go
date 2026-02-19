// Package types provides custom types for goConfy configuration fields.
//
// The primary type is [Duration], a wrapper around [time.Duration] that
// supports YAML marshaling and unmarshaling using Go's duration string
// format (e.g. "5m", "30s", "1h30m").
//
// Usage in config structs:
//
//	type Config struct {
//	    Timeout types.Duration `yaml:"timeout"`
//	}
//
// YAML:
//
//	timeout: 5m
package types
