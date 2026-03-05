package redact

import (
	"reflect"
	"strings"
)

// redactValue recursively walks the value and returns a map[string]any
// representation with sensitive fields replaced by "[REDACTED]".
func redactValue(v any, prefix string, pathSet map[string]bool, opts Options) any {
	if v == nil {
		return nil
	}
	rv := reflect.ValueOf(v)
	return redactReflect(rv, prefix, pathSet, opts)
}

func redactReflect(rv reflect.Value, prefix string, pathSet map[string]bool, opts Options) any {
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Struct:
		return redactStruct(rv, prefix, pathSet, opts)
	case reflect.Map:
		return redactMap(rv, prefix, pathSet, opts)
	case reflect.Slice, reflect.Array:
		return redactSlice(rv, prefix, pathSet, opts)
	default:
		return rv.Interface()
	}
}

func redactStruct(rv reflect.Value, prefix string, pathSet map[string]bool, opts Options) map[string]any {
	rt := rv.Type()
	result := make(map[string]any, rt.NumField())

	for i := range rt.NumField() {
		field := rt.Field(i)
		if !field.IsExported() {
			continue
		}

		name := fieldName(field)
		fullPath := joinPath(prefix, name)

		if field.Tag.Get("secret") == "true" || pathSet[fullPath] || (opts.ByConvention && IsConventionSecret(name)) {
			result[name] = redactedValue
			continue
		}

		result[name] = redactReflect(rv.Field(i), fullPath, pathSet, opts)
	}

	return result
}

func redactMap(rv reflect.Value, prefix string, pathSet map[string]bool, opts Options) map[string]any {
	result := make(map[string]any, rv.Len())
	iter := rv.MapRange()
	for iter.Next() {
		key := iter.Key().String()
		fullPath := joinPath(prefix, key)
		if pathSet[fullPath] || (opts.ByConvention && IsConventionSecret(key)) {
			result[key] = redactedValue
		} else {
			result[key] = redactReflect(iter.Value(), fullPath, pathSet, opts)
		}
	}
	return result
}

func redactSlice(rv reflect.Value, prefix string, pathSet map[string]bool, opts Options) []any {
	result := make([]any, rv.Len())
	for i := range rv.Len() {
		result[i] = redactReflect(rv.Index(i), prefix, pathSet, opts)
	}
	return result
}

func fieldName(f reflect.StructField) string {
	tag := f.Tag.Get("yaml")
	if tag == "" || tag == "-" {
		tag = f.Tag.Get("json")
	}
	if tag != "" && tag != "-" {
		parts := strings.SplitN(tag, ",", 2)
		if parts[0] != "" {
			return parts[0]
		}
	}
	return strings.ToLower(f.Name)
}

func joinPath(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}
