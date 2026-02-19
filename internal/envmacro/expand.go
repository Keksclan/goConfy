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
			if err := ExpandNode(child, opts); err != nil {
				return err
			}
		}
	case yaml.MappingNode:
		for i := 1; i < len(node.Content); i += 2 {
			if err := ExpandNode(node.Content[i], opts); err != nil {
				return err
			}
		}
	case yaml.SequenceNode:
		for _, child := range node.Content {
			if err := ExpandNode(child, opts); err != nil {
				return err
			}
		}
	case yaml.ScalarNode:
		return expandScalar(node, lookup, opts.Prefix, opts.AllowedKeys)
	}

	return nil
}

func expandScalar(node *yaml.Node, lookup LookupFunc, prefix string, allowedKeys []string) error {
	matches := EnvMacroRegex.FindStringSubmatch(node.Value)
	if matches == nil {
		return nil
	}

	key := matches[1]
	defaultVal := matches[2]

	if len(allowedKeys) > 0 && !slices.Contains(allowedKeys, key) {
		return fmt.Errorf("environment key %q is not in the allowed list", key)
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
		return fmt.Errorf("environment variable %q is not set and no default provided", key)
	}

	return nil
}

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
