package logic

import (
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// defaultSecretPaths are always redacted in YAML previews, in addition to
// any user-configured paths. These match the library's well-known secrets.
var defaultSecretPaths = []string{
	"redis.password",
	"auth.opaque.client_secret",
	"postgres.url",
}

// RedactYAML replaces scalar values at the given dot-paths (and the built-in
// defaults) with "******" in the provided YAML string. This is a best-effort
// redactor for human-readable YAML preview. For typed config, always prefer
// DumpRedactedJSON which uses the struct tags.
func RedactYAML(yamlStr string, extraPaths []string) (string, error) {
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(yamlStr), &node); err != nil {
		return "", err
	}

	// Merge default + user paths into a set.
	paths := make(map[string]bool, len(defaultSecretPaths)+len(extraPaths))
	for _, p := range defaultSecretPaths {
		paths[p] = true
	}
	for _, p := range extraPaths {
		paths[p] = true
	}

	root := &node
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		root = root.Content[0]
	}
	redactNode(root, "", paths)

	out, err := yaml.Marshal(&node)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

// redactNode walks a mapping node and replaces values whose dot-path is in
// the paths set with the redaction placeholder.
func redactNode(n *yaml.Node, prefix string, paths map[string]bool) {
	if n.Kind != yaml.MappingNode {
		return
	}
	for i := 0; i < len(n.Content)-1; i += 2 {
		key := n.Content[i].Value
		val := n.Content[i+1]

		fullPath := key
		if prefix != "" {
			fullPath = prefix + "." + key
		}

		if paths[fullPath] && val.Kind == yaml.ScalarNode {
			val.Value = "******"
			val.Tag = "!!str"
			continue
		}

		if val.Kind == yaml.MappingNode {
			redactNode(val, fullPath, paths)
		}
	}
}

// ContainsSecretPaths checks whether any of the default or extra secret paths
// exist in the given YAML. Returns true if at least one is found, indicating
// that the YAML preview should show a redaction warning.
func ContainsSecretPaths(yamlStr string, extraPaths []string) bool {
	allPaths := append(slices.Clone(defaultSecretPaths), extraPaths...)
	for _, p := range allPaths {
		// Quick heuristic: check if the leaf key appears in the YAML text.
		parts := strings.Split(p, ".")
		leaf := parts[len(parts)-1]
		if strings.Contains(yamlStr, leaf+":") {
			return true
		}
	}
	return false
}
