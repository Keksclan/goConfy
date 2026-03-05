package yamltemplate

import (
	"reflect"
	"strings"
)

// fieldInfo holds the extracted metadata from a struct field's tags.
type fieldInfo struct {
	// YAMLName is the key name for YAML output.
	YAMLName string
	// GoName is the original Go field name.
	GoName string
	// Default is the default value from `default:"..."` tag.
	Default string
	// HasDefault indicates whether a `default` tag was explicitly set.
	HasDefault bool
	// Desc is the description from `desc:"..."` tag.
	Desc string
	// Example is the example value from `example:"..."` tag.
	Example string
	// Options is the list of allowed values from `options:"..."` tag.
	Options string
	// Required indicates `required:"true"`.
	Required bool
	// Secret indicates `secret:"true"`.
	Secret bool
	// Env is the environment variable key from `env:"..."` tag.
	Env string
	// Sep is the separator for slice env parsing from `sep:"..."` tag.
	Sep string
	// Kind is the reflect.Kind of the field's type.
	Kind reflect.Kind
	// Type is the reflect.Type of the field.
	Type reflect.Type
	// IsStruct indicates the field is a struct (not a special type like Duration).
	IsStruct bool
	// Children holds nested field info for struct fields.
	Children []fieldInfo
}

// extractFields uses reflection to walk the struct type and extract field metadata.
func extractFields(t reflect.Type) []fieldInfo {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	var fields []fieldInfo
	for i := range t.NumField() {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		yamlTag := f.Tag.Get("yaml")
		if yamlTag == "-" {
			continue
		}

		yamlName := yamlKeyName(f, yamlTag)
		if yamlName == "" {
			continue
		}

		info := fieldInfo{
			YAMLName: yamlName,
			GoName:   f.Name,
			Kind:     f.Type.Kind(),
			Type:     f.Type,
		}

		// Extract tags.
		if def, ok := f.Tag.Lookup("default"); ok {
			info.Default = def
			info.HasDefault = true
		}
		info.Desc = f.Tag.Get("desc")
		info.Example = f.Tag.Get("example")
		info.Options = f.Tag.Get("options")
		info.Required = f.Tag.Get("required") == "true"
		info.Secret = f.Tag.Get("secret") == "true"
		info.Env = f.Tag.Get("env")
		info.Sep = f.Tag.Get("sep")

		// Check if this is a nested struct (not a special type).
		ft := f.Type
		if ft.Kind() == reflect.Ptr {
			ft = ft.Elem()
		}
		if ft.Kind() == reflect.Struct && !isSpecialType(ft) {
			info.IsStruct = true
			info.Children = extractFields(ft)
		}

		fields = append(fields, info)
	}
	return fields
}

// yamlKeyName extracts the YAML key name from a struct field and its yaml tag.
func yamlKeyName(f reflect.StructField, tag string) string {
	if tag != "" {
		parts := strings.SplitN(tag, ",", 2)
		if parts[0] != "" {
			return parts[0]
		}
	}
	return strings.ToLower(f.Name)
}

// isSpecialType returns true for types that should be treated as scalars
// rather than recursed into (e.g. time.Time, types.Duration).
func isSpecialType(t reflect.Type) bool {
	// Check if the type implements yaml.Marshaler (has MarshalYAML method).
	for i := range t.NumMethod() {
		if t.Method(i).Name == "MarshalYAML" {
			return true
		}
	}
	// Also check pointer receiver.
	pt := reflect.PointerTo(t)
	for i := range pt.NumMethod() {
		if pt.Method(i).Name == "MarshalYAML" {
			return true
		}
	}
	return false
}
