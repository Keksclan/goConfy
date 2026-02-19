// Package redact provides deep-copy redaction of configuration structs.
//
// It walks a struct value using reflection and replaces sensitive fields with
// the placeholder "[REDACTED]". Fields are considered sensitive if they carry
// the struct tag `secret:"true"` or if their dot-separated path appears in
// an explicit redaction path list.
//
// The redacted output is a map[string]any suitable for JSON serialization,
// ensuring that secret values never leak into logs or debug output.
//
// This is an internal package used by the goconfy redaction helpers.
package redact
