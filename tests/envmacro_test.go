package tests

import (
	"testing"

	"github.com/keksclan/goConfy/internal/envmacro"
	"gopkg.in/yaml.v3"
)

func TestEnvMacroPattern(t *testing.T) {
	tests := []struct {
		input string
		match bool
	}{
		{"{ENV:PORT:8080}", true},
		{"{ENV:HOST}", true},
		{"{ENV:DB_NAME:mydb}", true},
		{"{ENV:A1_B2:val}", true},
		// Must NOT match
		{"http://{ENV:HOST}/x", false},
		{"prefix{ENV:PORT:8080}", false},
		{"{ENV:PORT:8080}suffix", false},
		{"${PORT}", false},
		{"{ENV:lowercase}", false},
		{"plaintext", false},
		{"", false},
	}

	for _, tt := range tests {
		matched := envmacro.EnvMacroRegex.MatchString(tt.input)
		if matched != tt.match {
			t.Errorf("input=%q: expected match=%v, got %v", tt.input, tt.match, matched)
		}
	}
}

func TestExpandNodeBasic(t *testing.T) {
	input := `
host: "{ENV:MYHOST:defaulthost}"
port: "{ENV:MYPORT:3000}"
`
	node := parseYAML(t, input)

	lookup := func(key string) (string, bool) {
		if key == "MYHOST" {
			return "resolved-host", true
		}
		return "", false
	}

	err := envmacro.ExpandNode(node, envmacro.ExpandOptions{Lookup: lookup})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	values := extractScalars(node)
	if values["host"] != "resolved-host" {
		t.Errorf("expected host=resolved-host, got %q", values["host"])
	}
	if values["port"] != "3000" {
		t.Errorf("expected port=3000 (default), got %q", values["port"])
	}
}

func TestInlineMacroNotExpanded(t *testing.T) {
	input := `
url: "http://{ENV:HOST}/api"
`
	node := parseYAML(t, input)

	err := envmacro.ExpandNode(node, envmacro.ExpandOptions{
		Lookup: func(key string) (string, bool) {
			return "shouldnotappear", true
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	values := extractScalars(node)
	if values["url"] != "http://{ENV:HOST}/api" {
		t.Errorf("inline macro should NOT be expanded, got %q", values["url"])
	}
}

func TestExpandNodeAllowedKeys(t *testing.T) {
	input := `
host: "{ENV:ALLOWED_KEY:default}"
`
	node := parseYAML(t, input)

	err := envmacro.ExpandNode(node, envmacro.ExpandOptions{
		Lookup:      func(string) (string, bool) { return "", false },
		AllowedKeys: []string{"OTHER_KEY"},
	})
	if err == nil {
		t.Fatal("expected error for disallowed key")
	}

	if ife, ok := err.(interface {
		GetPath() string
		GetLine() int
	}); ok {
		if ife.GetPath() != "host" {
			t.Errorf("expected path 'host', got %q", ife.GetPath())
		}
		if ife.GetLine() != 2 {
			t.Errorf("expected line 2, got %d", ife.GetLine())
		}
	} else {
		t.Errorf("expected error to provide Path and Line, got %T: %v", err, err)
	}
}

func TestExpandNodeDeepPath(t *testing.T) {
	input := `
server:
  addr:
    host: "{ENV:MISSING}"
`
	node := parseYAML(t, input)

	err := envmacro.ExpandNode(node, envmacro.ExpandOptions{
		Lookup: func(string) (string, bool) { return "", false },
	})
	if err == nil {
		t.Fatal("expected error for missing env var")
	}

	if ife, ok := err.(interface {
		GetPath() string
	}); ok {
		if ife.GetPath() != "server.addr.host" {
			t.Errorf("expected path 'server.addr.host', got %q", ife.GetPath())
		}
	} else {
		t.Errorf("expected error to provide Path, got %v", err)
	}
}

func TestExpandNodeWithPrefix(t *testing.T) {
	input := `
host: "{ENV:HOST:fallback}"
`
	node := parseYAML(t, input)

	err := envmacro.ExpandNode(node, envmacro.ExpandOptions{
		Lookup: func(key string) (string, bool) {
			if key == "APP_HOST" {
				return "prefixed-host", true
			}
			return "", false
		},
		Prefix: "APP_",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	values := extractScalars(node)
	if values["host"] != "prefixed-host" {
		t.Errorf("expected host=prefixed-host, got %q", values["host"])
	}
}

func parseYAML(t *testing.T, input string) *yaml.Node {
	t.Helper()
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(input), &node); err != nil {
		t.Fatalf("failed to parse YAML: %v", err)
	}
	return &node
}

func extractScalars(node *yaml.Node) map[string]string {
	result := make(map[string]string)
	var walk func(*yaml.Node)
	walk = func(n *yaml.Node) {
		if n.Kind == yaml.MappingNode {
			for i := 0; i < len(n.Content)-1; i += 2 {
				key := n.Content[i].Value
				val := n.Content[i+1]
				if val.Kind == yaml.ScalarNode {
					result[key] = val.Value
				} else {
					walk(val)
				}
			}
		}
		if n.Kind == yaml.DocumentNode {
			for _, c := range n.Content {
				walk(c)
			}
		}
	}
	walk(node)
	return result
}
