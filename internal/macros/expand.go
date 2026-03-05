package macros

import (
	"fmt"
	"os"
	"regexp"
	"slices"
	"strings"

	"gopkg.in/yaml.v3"
)

// FileMacroRegex is the regex for file macros.
// Pattern: {FILE:/path/to/file}, {FILE:C:\path} or {FILE:/path:default}
var FileMacroRegex = regexp.MustCompile(`^\{FILE:([^}]+)\}$`)

// EnvMacroRegex is the regex for environment macros.
// Pattern: {ENV:KEY} or {ENV:KEY:default}
var EnvMacroRegex = regexp.MustCompile(`^\{ENV:([A-Z0-9_]+)(?::([^}]*))?\}$`)

// ExpandOptions controls macro expansion behavior.
type ExpandOptions struct {
	LookupEnv   func(string) (string, bool)
	EnvPrefix   string
	AllowedKeys []string
	OnExpand    func(path, key, value string, source string)
}

// ExpandNode recursively walks a yaml.Node tree and expands both
// environment {ENV:KEY} and file {FILE:/path} macros on scalar values.
func ExpandNode(node *yaml.Node, opts ExpandOptions) error {
	return expandNode(node, opts, "")
}

func expandNode(node *yaml.Node, opts ExpandOptions, path string) error {
	if node == nil {
		return nil
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
			var newPath string
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
		return expandScalar(node, opts, path)
	}

	return nil
}

func expandScalar(node *yaml.Node, opts ExpandOptions, path string) error {
	// Try FILE macro
	if matches := FileMacroRegex.FindStringSubmatch(node.Value); matches != nil {
		filePath, defaultVal, hasDefault := parseFileMacro(matches[1])

		data, err := os.ReadFile(filePath)
		if err != nil {
			if hasDefault {
				node.Value = defaultVal
				node.Tag = ""
				node.Style = 0
				if opts.OnExpand != nil {
					opts.OnExpand(path, filePath, defaultVal, "default")
				}
				return nil
			}
			return &FieldError{
				Path:    path,
				Line:    node.Line,
				Column:  node.Column,
				Message: fmt.Sprintf("failed to read file %q: %v", filePath, err),
			}
		}

		val := strings.TrimSpace(string(data))
		node.Value = val
		node.Tag = ""
		node.Style = 0
		if opts.OnExpand != nil {
			opts.OnExpand(path, filePath, val, "file")
		}
		return nil
	}

	// Try ENV macro
	if matches := EnvMacroRegex.FindStringSubmatch(node.Value); matches != nil {
		key := matches[1]
		defaultVal := matches[2]
		colons := 0
		for _, ch := range matches[0] {
			if ch == ':' {
				colons++
			}
		}
		hasDefault := colons >= 2

		if len(opts.AllowedKeys) > 0 && !slices.Contains(opts.AllowedKeys, key) {
			if hasDefault {
				node.Value = defaultVal
				node.Tag = ""
				node.Style = 0
				if opts.OnExpand != nil {
					opts.OnExpand(path, key, defaultVal, "default")
				}
				return nil
			}
			return &FieldError{
				Path:    path,
				Line:    node.Line,
				Column:  node.Column,
				Message: fmt.Sprintf("environment key %q is not in the allowed list", key),
			}
		}

		lookupKey := key
		if opts.EnvPrefix != "" {
			lookupKey = opts.EnvPrefix + key
		}

		lookup := opts.LookupEnv
		if lookup == nil {
			lookup = os.LookupEnv
		}

		if val, ok := lookup(lookupKey); ok {
			node.Value = val
			node.Tag = ""
			node.Style = 0
			if opts.OnExpand != nil {
				opts.OnExpand(path, key, val, "env")
			}
		} else if hasDefault {
			node.Value = defaultVal
			node.Tag = ""
			node.Style = 0
			if opts.OnExpand != nil {
				opts.OnExpand(path, key, defaultVal, "default")
			}
		} else {
			return &FieldError{
				Path:    path,
				Line:    node.Line,
				Column:  node.Column,
				Message: fmt.Sprintf("environment variable %q is not set and no default provided", key),
			}
		}
		return nil
	}

	return nil
}

type FieldError struct {
	Path    string
	Line    int
	Column  int
	Message string
}

func (e *FieldError) Error() string {
	if e.Line > 0 {
		if e.Column > 0 {
			return fmt.Sprintf("path %q: line %d, col %d: %s", e.Path, e.Line, e.Column, e.Message)
		}
		return fmt.Sprintf("path %q: line %d: %s", e.Path, e.Line, e.Message)
	}
	return fmt.Sprintf("path %q: %s", e.Path, e.Message)
}

func (e *FieldError) GetPath() string { return e.Path }
func (e *FieldError) GetLine() int    { return e.Line }
func (e *FieldError) GetColumn() int  { return e.Column }

func parseFileMacro(content string) (filePath, defaultVal string, hasDefault bool) {
	lastIdx := strings.LastIndex(content, ":")
	if lastIdx == -1 {
		return content, "", false
	}

	if lastIdx == 1 && len(content) > 1 &&
		((content[0] >= 'a' && content[0] <= 'z') || (content[0] >= 'A' && content[0] <= 'Z')) {
		// Windows drive letter C: at the start, and it's the only colon
		return content, "", false
	}

	return content[:lastIdx], content[lastIdx+1:], true
}
