package yamlparse

import (
	"bytes"
	"fmt"
	"io"

	"gopkg.in/yaml.v3"
)

// ParseBytes parses YAML bytes into a yaml.Node.
func ParseBytes(data []byte) (*yaml.Node, error) {
	return ParseReader(bytes.NewReader(data))
}

// ParseReader parses YAML from a reader into a yaml.Node.
func ParseReader(r io.Reader) (*yaml.Node, error) {
	var node yaml.Node
	dec := yaml.NewDecoder(r)
	if err := dec.Decode(&node); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to parse YAML: %w", err)
	}
	return &node, nil
}
