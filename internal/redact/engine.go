package redact

const redactedValue = "[REDACTED]"

// Options controls which fields are redacted.
type Options struct {
	// Paths is a list of dot-separated paths to redact (e.g., "db.password").
	Paths []string
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
