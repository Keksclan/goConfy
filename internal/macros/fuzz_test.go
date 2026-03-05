package macros

import (
	"testing"

	"gopkg.in/yaml.v3"
)

func FuzzExpandNode(f *testing.F) {
	seeds := []string{
		"{ENV:VAR}",
		"{ENV:VAR:default}",
		"{FILE:/tmp/test}",
		"{FILE:/tmp/test:default}",
		"simple string",
		"key: {ENV:VAR}",
		"nested: {FILE:path}",
		"{ENV:INVALID-KEY}",
		"{ENV:VAR:with:colons}",
		"{FILE:path:with:colons}",
		"{INVALID:MACRO}",
		"{}",
		"{:}",
		"{{ENV:VAR}}",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	f.Fuzz(func(t *testing.T, input string) {
		var node yaml.Node
		err := yaml.Unmarshal([]byte(input), &node)
		if err != nil {
			// Ignore invalid YAML inputs as the fuzzer should focus on macro expansion
			return
		}

		opts := ExpandOptions{
			LookupEnv: func(s string) (string, bool) {
				if s == "VAR" || s == "PREFIX_VAR" {
					return "expanded_value", true
				}
				return "", false
			},
			EnvPrefix:   "PREFIX_",
			AllowedKeys: []string{"VAR", "OTHER"},
			OnExpand:    func(path, key, value, source string) {},
		}

		// We expect ExpandNode never to panic
		_ = ExpandNode(&node, opts)
	})
}
