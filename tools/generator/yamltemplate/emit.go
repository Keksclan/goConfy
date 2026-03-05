// Package yamltemplate generates YAML config templates from Go struct types
// using reflection and struct tags.
package yamltemplate

import (
	"fmt"
	"reflect"
	"strings"
)

// Options controls the YAML template generation.
type Options struct {
	// Profile, if non-empty, appends a profiles skeleton block.
	Profile string
	// DotEnv, if true, also generates .env content.
	DotEnv bool
}

// Generate produces a YAML template string from a config struct pointer.
// The value must be a pointer to a struct.
func Generate(v any, opts Options) (string, error) {
	t := reflect.TypeOf(v)
	if t == nil {
		return "", fmt.Errorf("yamltemplate: nil value")
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("yamltemplate: expected struct, got %s", t.Kind())
	}

	fields := extractFields(t)
	var buf strings.Builder
	emitFields(&buf, fields, 0)

	if opts.Profile != "" {
		emitProfileSkeleton(&buf, opts.Profile, fields)
	}

	return buf.String(), nil
}

// GenerateDotEnv produces a sample .env file content from struct tags.
func GenerateDotEnv(v any) (string, error) {
	t := reflect.TypeOf(v)
	if t == nil {
		return "", fmt.Errorf("yamltemplate: nil value")
	}
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return "", fmt.Errorf("yamltemplate: expected struct, got %s", t.Kind())
	}

	fields := extractFields(t)
	var buf strings.Builder
	emitDotEnvFields(&buf, fields)
	return buf.String(), nil
}

// emitFields writes YAML key-value pairs with comments at the given indentation level.
func emitFields(buf *strings.Builder, fields []fieldInfo, indent int) {
	prefix := strings.Repeat("  ", indent)

	for i, fi := range fields {
		if i > 0 {
			// Add blank line between top-level fields for readability.
			if indent == 0 {
				buf.WriteString("\n")
			}
		}

		// Emit comment block.
		comment := buildComment(fi)
		if comment != "" {
			for _, line := range strings.Split(comment, "\n") {
				buf.WriteString(prefix)
				buf.WriteString("# ")
				buf.WriteString(line)
				buf.WriteString("\n")
			}
		}

		if fi.IsStruct {
			buf.WriteString(prefix)
			buf.WriteString(fi.YAMLName)
			buf.WriteString(":\n")
			emitFields(buf, fi.Children, indent+1)
		} else if fi.Kind == reflect.Slice || fi.Kind == reflect.Array {
			emitSliceField(buf, fi, prefix)
		} else {
			val := fieldValue(fi)
			buf.WriteString(prefix)
			buf.WriteString(fi.YAMLName)
			buf.WriteString(": ")
			buf.WriteString(formatYAMLValue(val))
			buf.WriteString("\n")
		}
	}
}

// emitSliceField writes a YAML list field.
func emitSliceField(buf *strings.Builder, fi fieldInfo, prefix string) {
	if fi.Env != "" {
		// Env-sourced slices use macro format.
		val := fieldValue(fi)
		buf.WriteString(prefix)
		buf.WriteString(fi.YAMLName)
		buf.WriteString(": ")
		buf.WriteString(formatYAMLValue(val))
		buf.WriteString("\n")
		return
	}

	buf.WriteString(prefix)
	buf.WriteString(fi.YAMLName)
	buf.WriteString(":\n")

	if fi.Example != "" {
		// Show example items.
		for _, item := range strings.Split(fi.Example, ",") {
			buf.WriteString(prefix)
			buf.WriteString("  - ")
			buf.WriteString(strings.TrimSpace(item))
			buf.WriteString("\n")
		}
	} else {
		// Empty placeholder.
		buf.WriteString(prefix)
		buf.WriteString("  # - item\n")
	}
}

// formatYAMLValue wraps string values in quotes when appropriate.
func formatYAMLValue(val string) string {
	if val == "" {
		return `""`
	}
	// Macro values and special strings get quoted.
	if strings.HasPrefix(val, "{ENV:") {
		return `"` + val + `"`
	}
	// Booleans and numbers can stay unquoted.
	switch val {
	case "true", "false":
		return val
	}
	// Check if it looks numeric.
	if isNumeric(val) {
		return val
	}
	return `"` + val + `"`
}

// isNumeric returns true if the string represents a number.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	start := 0
	if s[0] == '-' || s[0] == '+' {
		start = 1
	}
	dotSeen := false
	for i := start; i < len(s); i++ {
		if s[i] == '.' && !dotSeen {
			dotSeen = true
			continue
		}
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return start < len(s)
}

// emitProfileSkeleton appends a profiles block at the end of the YAML.
func emitProfileSkeleton(buf *strings.Builder, profile string, fields []fieldInfo) {
	buf.WriteString("\n# Profile overrides\nprofiles:\n")

	profiles := []string{profile}
	// Add common profiles if the specified one isn't already there.
	for _, p := range []string{"dev", "staging", "prod"} {
		if p != profile {
			profiles = append(profiles, p)
		}
	}

	for _, p := range profiles {
		buf.WriteString("  ")
		buf.WriteString(p)
		buf.WriteString(":\n")
		// Add a couple example overrides from top-level scalar fields.
		count := 0
		for _, fi := range fields {
			if !fi.IsStruct && fi.Kind != reflect.Slice && fi.Kind != reflect.Array {
				buf.WriteString("    # ")
				buf.WriteString(fi.YAMLName)
				buf.WriteString(": ")
				buf.WriteString(formatYAMLValue(fieldValue(fi)))
				buf.WriteString("\n")
				count++
				if count >= 2 {
					break
				}
			}
		}
	}
}

// emitDotEnvFields writes .env file content for all env-tagged fields.
func emitDotEnvFields(buf *strings.Builder, fields []fieldInfo) {
	for _, fi := range fields {
		if fi.IsStruct {
			emitDotEnvFields(buf, fi.Children)
			continue
		}
		if fi.Env == "" {
			continue
		}

		// Comment with description.
		if fi.Desc != "" {
			buf.WriteString("# ")
			buf.WriteString(fi.Desc)
			buf.WriteString("\n")
		}

		def := defaultValue(fi)
		if fi.Secret {
			buf.WriteString(fi.Env)
			buf.WriteString("=\n")
		} else {
			buf.WriteString(fi.Env)
			buf.WriteString("=")
			buf.WriteString(def)
			buf.WriteString("\n")
		}
	}
}
