package config

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"

	"github.com/keksclan/goConfy"
	"gopkg.in/yaml.v3"
)

func TestConfigSync(t *testing.T) {
	// Mock DB_PASSWORD for loader (required in example.yml)
	t.Setenv("DB_PASSWORD", "test-pass")

	// 1. Ensure config.example.yml can be loaded into Config struct
	// We use Strict mode (which is default in Load) to ensure no unknown keys.
	_, err := goconfy.Load[Config](
		goconfy.WithFile("../../config.example.yml"),
		goconfy.WithEnableProfiles(true),
	)
	if err != nil {
		t.Fatalf("failed to load config.example.yml: %v", err)
	}

	// 2. Ensure every field in the Config struct is present in the example.yml
	data, err := os.ReadFile("../../config.example.yml")
	if err != nil {
		t.Fatalf("failed to read config.example.yml: %v", err)
	}

	var node yaml.Node
	if err := yaml.Unmarshal(data, &node); err != nil {
		t.Fatalf("failed to unmarshal yaml: %v", err)
	}

	checkFields(t, reflect.TypeFor[Config](), &node, "")

	// 3. Ensure JSON Schema is present and top-level keys match
	checkSchema(t)
}

func checkSchema(t *testing.T) {
	data, err := os.ReadFile("../../schema/config.schema.json")
	if err != nil {
		t.Fatalf("failed to read schema: %v", err)
	}

	var schema struct {
		Properties map[string]any `json:"properties"`
	}
	if err := json.Unmarshal(data, &schema); err != nil {
		t.Fatalf("failed to unmarshal schema: %v", err)
	}

	rt := reflect.TypeFor[Config]()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "-" || yamlTag == "" {
			continue
		}
		if _, ok := schema.Properties[yamlTag]; !ok {
			t.Errorf("field %q missing in JSON schema properties", yamlTag)
		}
	}
}

func checkFields(t *testing.T, rt reflect.Type, node *yaml.Node, path string) {
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}

	if rt.Kind() != reflect.Struct {
		return
	}

	// Get mapping node
	var mapping *yaml.Node
	if node.Kind == yaml.DocumentNode {
		mapping = node.Content[0]
	} else {
		mapping = node
	}

	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		yamlTag := field.Tag.Get("yaml")
		if yamlTag == "-" || yamlTag == "" {
			continue
		}

		key := yamlTag
		fieldPath := key
		if path != "" {
			fieldPath = path + "." + key
		}

		// Check if key exists in mapping
		found := false
		var valNode *yaml.Node
		for j := 0; j < len(mapping.Content)-1; j += 2 {
			if mapping.Content[j].Value == key {
				found = true
				valNode = mapping.Content[j+1]
				break
			}
		}

		if !found {
			t.Errorf("field %q (Go: %s) missing in config.example.yml", fieldPath, field.Name)
			continue
		}

		// Recurse for structs
		if field.Type.Kind() == reflect.Struct && field.Type.String() != "types.Duration" {
			checkFields(t, field.Type, valNode, fieldPath)
		}
	}
}

func TestNoHiddenConfig(t *testing.T) {
	// This test would ideally scan the codebase for os.Getenv() calls
	// that are not represented in the config.
	// For now, we manually ensure our canonical Config covers common needs.
}
