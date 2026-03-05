package yamltemplate

import (
	"reflect"
)

// defaultValue returns the appropriate default value string for a field.
// It considers the field's tags, type, and whether it's a secret.
func defaultValue(fi fieldInfo) string {
	// Secrets: never output real default values.
	if fi.Secret {
		return ""
	}

	// Explicit default tag takes priority.
	if fi.HasDefault {
		return fi.Default
	}

	// Sensible zero-value defaults based on type kind.
	return zeroDefault(fi.Kind)
}

// zeroDefault returns a sensible zero-value string for common Go kinds.
func zeroDefault(k reflect.Kind) string {
	switch k {
	case reflect.Bool:
		return "false"
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return "0"
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "0"
	case reflect.Float32, reflect.Float64:
		return "0"
	case reflect.String:
		return ""
	default:
		return ""
	}
}

// fieldValue returns the YAML value string for a field, applying macro expansion
// for env-sourced fields and redaction for secrets.
func fieldValue(fi fieldInfo) string {
	if fi.Env != "" {
		def := defaultValue(fi)
		if fi.Secret {
			// Never embed secret defaults in macros.
			return "{ENV:" + fi.Env + ":}"
		}
		return "{ENV:" + fi.Env + ":" + def + "}"
	}
	return defaultValue(fi)
}
