package redact

import (
	"reflect"
	"strings"
)

const redactedValue = "[REDACTED]"

// Options controls which fields are redacted.
type Options struct {
	// Paths is a list of dot-separated paths to redact (e.g., "db.password").
	Paths []string

	// ByConvention enables automatic redaction based on field names
	// (e.g., password, secret, token, key, private).
	ByConvention bool
}

// IsSecret checks if a given path in a struct type is marked as secret.
// It considers struct tags, explicit paths, and name conventions if enabled.
// It expects a dot-separated path (e.g., "db.password").
func IsSecret(t reflect.Type, path string, opts Options) bool {
	if t == nil {
		return false
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}

	pathSet := make(map[string]bool, len(opts.Paths))
	for _, p := range opts.Paths {
		pathSet[p] = true
	}

	parts := strings.Split(path, ".")
	curr := t
	for i, part := range parts {
		found := false
		for j := range curr.NumField() {
			field := curr.Field(j)
			name := fieldName(field)
			if name == part {
				fullPath := strings.Join(parts[:i+1], ".")

				if field.Tag.Get("secret") == "true" || pathSet[fullPath] || (opts.ByConvention && IsConventionSecret(name)) {
					return true
				}
				if i == len(parts)-1 {
					return false
				}
				curr = field.Type
				for curr.Kind() == reflect.Ptr || curr.Kind() == reflect.Slice || curr.Kind() == reflect.Array {
					curr = curr.Elem()
				}
				if curr.Kind() != reflect.Struct {
					return false
				}
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return false
}

// StripIndices removes all array/slice indices from a path (e.g., "foo[0].bar" -> "foo.bar").
func StripIndices(path string) string {
	if !strings.Contains(path, "[") {
		return path
	}
	var b strings.Builder
	b.Grow(len(path))
	skip := false
	for i := 0; i < len(path); i++ {
		if path[i] == '[' {
			skip = true
			continue
		}
		if path[i] == ']' {
			skip = false
			continue
		}
		if !skip {
			b.WriteByte(path[i])
		}
	}
	return b.String()
}

var conventionKeywords = []string{"password", "secret", "token", "key", "private"}

// IsConventionSecret returns true if the name matches a common secret convention.
func IsConventionSecret(name string) bool {
	name = strings.ToLower(name)
	for _, kw := range conventionKeywords {
		if strings.Contains(name, kw) {
			return true
		}
	}
	return false
}

// Apply creates a redacted copy of the given value, replacing sensitive fields
// with "[REDACTED]". It checks both the `secret:"true"` struct tag and
// explicit dot-path configuration.
func Apply(v any, opts Options) any {
	pathSet := make(map[string]bool, len(opts.Paths))
	for _, p := range opts.Paths {
		pathSet[p] = true
	}
	return redactValue(v, "", pathSet, opts)
}
