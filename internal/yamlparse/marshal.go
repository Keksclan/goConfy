package yamlparse

import (
	"bytes"
	"fmt"

	"gopkg.in/yaml.v3"
)

// MarshalNode marshals a yaml.Node back to bytes.
func MarshalNode(node *yaml.Node) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(node); err != nil {
		return nil, fmt.Errorf("failed to marshal YAML node: %w", err)
	}
	return buf.Bytes(), nil
}
