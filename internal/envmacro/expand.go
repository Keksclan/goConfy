package envmacro

import (
	"fmt"
	"slices"

	"gopkg.in/yaml.v3"
)

// ExpandOptions controls macro expansion behavior.
type ExpandOptions struct {
	Lookup      LookupFunc
	Prefix      string
	AllowedKeys []string
}

// ExpandNode recursively walks a yaml.Node tree and expands environment macros
// on scalar values that exactly match the {ENV:KEY:default} pattern.
func ExpandNode(node *yaml.Node, opts ExpandOptions) error {
	return expandNode(node, opts, "")
}

func expandNode(node *yaml.Node, opts ExpandOptions, path string) error {
	if node == nil {
		return nil
	}

	lookup := opts.Lookup
	if lookup == nil {
		lookup = DefaultLookup
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			if err := expandNode(child, opts, path); err != nil {
				return err
			}
		}
	case yaml.MappingNode:
		for i := 0; i < len(node.Content)-1; i += 2 {
			keyNode := node.Content[i]
			valNode := node.Content[i+1]
			newPath := path
			if path == "" {
				newPath = keyNode.Value
			} else {
				newPath = path + "." + keyNode.Value
			}
			if err := expandNode(valNode, opts, newPath); err != nil {
				return err
			}
		}
	case yaml.SequenceNode:
		for i, child := range node.Content {
			newPath := fmt.Sprintf("%s[%d]", path, i)
			if err := expandNode(child, opts, newPath); err != nil {
				return err
			}
		}
	case yaml.ScalarNode:
		return expandScalar(node, lookup, opts.Prefix, opts.AllowedKeys, path)
	}

	return nil
}

func expandScalar(node *yaml.Node, lookup LookupFunc, prefix string, allowedKeys []string, path string) error {
	matches := EnvMacroRegex.FindStringSubmatch(node.Value)
	if matches == nil {
		return nil
	}

	key := matches[1]
	defaultVal := matches[2]

	if len(allowedKeys) > 0 && !slices.Contains(allowedKeys, key) {
		return &fieldError{
			Path:    path,
			Line:    node.Line,
			Column:  node.Column,
			Message: fmt.Sprintf("environment key %q is not in the allowed list", key),
		}
	}

	lookupKey := key
	if prefix != "" {
		lookupKey = prefix + key
	}

	if val, ok := lookup(lookupKey); ok {
		node.Value = val
		node.Tag = ""
		node.Style = 0
	} else if hasDefault(matches) {
		node.Value = defaultVal
		node.Tag = ""
		node.Style = 0
	} else {
		return &fieldError{
			Path:    path,
			Line:    node.Line,
			Column:  node.Column,
			Message: fmt.Sprintf("environment variable %q is not set and no default provided", key),
		}
	}

	return nil
}

// fieldError matches the interface needed by goconfy.FieldError
// but avoids circular dependency.
type fieldError struct {
	Path    string
	Line    int
	Column  int
	Message string
}

func (e *fieldError) Error() string {
	return fmt.Sprintf("path %q: line %d, col %d: %s", e.Path, e.Line, e.Column, e.Message)
}

func (e *fieldError) GetPath() string { return e.Path }
func (e *fieldError) GetLine() int    { return e.Line }
func (e *fieldError) GetColumn() int  { return e.Column }

// hasDefault returns true if the regex match contains the optional default part
// (i.e. the colon separator was present, even if the default value is empty).
func hasDefault(matches []string) bool {
	if len(matches) < 3 {
		return false
	}
	// Count colons: {ENV:KEY} has 1, {ENV:KEY:} or {ENV:KEY:val} has 2+.
	colons := 0
	for _, ch := range matches[0] {
		if ch == ':' {
			colons++
		}
	}
	return colons >= 2
}
