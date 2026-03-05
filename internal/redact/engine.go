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
}

// IsSecret checks if a given path in a struct type is marked as secret.
func IsSecret(t reflect.Type, path string) bool {
	if t == nil {
		return false
	}
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Interface {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return false
	}

	parts := strings.Split(path, ".")
	curr := t
	for i, part := range parts {
		found := false
		for j := 0; j < curr.NumField(); j++ {
			field := curr.Field(j)
			name := fieldName(field)
			if name == part {
				if field.Tag.Get("secret") == "true" {
					return true
				}
				if i == len(parts)-1 {
					return false
				}
				curr = field.Type
				for curr.Kind() == reflect.Ptr || curr.Kind() == reflect.Interface {
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

// Apply creates a redacted copy of the given value, replacing sensitive fields
// with "[REDACTED]". It checks both the `secret:"true"` struct tag and
// explicit dot-path configuration.
func Apply(v any, opts Options) any {
	pathSet := make(map[string]bool, len(opts.Paths))
	for _, p := range opts.Paths {
		pathSet[p] = true
	}
	return redactValue(v, "", pathSet)
}
