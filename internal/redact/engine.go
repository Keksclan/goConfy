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
		// Strip indices like [0] from the part to match struct field names
		cleanPart := part
		if idx := strings.IndexByte(part, '['); idx != -1 {
			cleanPart = part[:idx]
		}

		found := false
		for j := 0; j < curr.NumField(); j++ {
			field := curr.Field(j)
			name := fieldName(field)
			if name == cleanPart {
				// Also strip indices from parts used to build fullPath for pathSet match
				fullPathParts := make([]string, i+1)
				for k := 0; k <= i; k++ {
					p := parts[k]
					if idx := strings.IndexByte(p, '['); idx != -1 {
						p = p[:idx]
					}
					fullPathParts[k] = p
				}
				fullPath := strings.Join(fullPathParts, ".")

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
