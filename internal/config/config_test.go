package config

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
	for i := range rt.NumField() {
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

	for i := range rt.NumField() {
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
	// 1. Get all allowed env keys from the Config struct
	allowedKeys := make(map[string]struct{})
	collectEnvTags(reflect.TypeFor[Config](), allowedKeys)

	// 2. Scan the repository for os.Getenv calls
	// We start from the project root (relative to this test file)
	root := "../../"
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor directory entirely
		if info.IsDir() && info.Name() == "vendor" {
			return filepath.SkipDir
		}

		// Skip directories and non-Go files
		if info.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}

		// Skip test files to avoid noise from test setup
		if strings.HasSuffix(path, "_test.go") {
			return nil
		}

		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil // Skip files that don't parse
		}

		ast.Inspect(f, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			// Check if it's os.Getenv
			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			x, ok := sel.X.(*ast.Ident)
			if !ok || x.Name != "os" || sel.Sel.Name != "Getenv" {
				return true
			}

			// We found a call to os.Getenv.
			// Now check the first argument.
			if len(call.Args) == 0 {
				return true
			}

			arg := call.Args[0]
			lit, ok := arg.(*ast.BasicLit)
			if !ok || lit.Kind != token.STRING {
				// Dynamic call (e.g. os.Getenv(envVar)), we skip it as it's
				// likely used in infrastructure code like profiles.SelectProfile.
				return true
			}

			// It's a string literal like "DB_PASSWORD"
			envKey := strings.Trim(lit.Value, "\"")
			if _, ok := allowedKeys[envKey]; !ok {
				t.Errorf("Hidden configuration found: os.Getenv(%q) at %s:%d. "+
					"This key is not represented in the canonical Config struct.",
					envKey, path, fset.Position(lit.Pos()).Line)
			}

			return true
		})

		return nil
	})

	if err != nil {
		t.Fatalf("failed to walk repository: %v", err)
	}
}

func collectEnvTags(rt reflect.Type, tags map[string]struct{}) {
	if rt.Kind() == reflect.Ptr {
		rt = rt.Elem()
	}
	if rt.Kind() != reflect.Struct {
		return
	}

	for i := range rt.NumField() {
		field := rt.Field(i)
		if tag := field.Tag.Get("env"); tag != "" {
			tags[tag] = struct{}{}
		}
		// Recursively collect from nested structs
		collectEnvTags(field.Type, tags)
	}
}
