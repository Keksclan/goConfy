package envmacro

import (
	"fmt"

	"github.com/keksclan/goConfy/internal/macros"
	"gopkg.in/yaml.v3"
)

// ExpandOptions controls macro expansion behavior.
type ExpandOptions struct {
	Lookup      LookupFunc
	Prefix      string
	AllowedKeys []string
	OnExpand    func(path, key, value string, source string)
}

// ExpandNode recursively walks a yaml.Node tree and expands environment macros
// on scalar values that exactly match the {ENV:KEY:default} pattern.
func ExpandNode(node *yaml.Node, opts ExpandOptions) error {
	return macros.ExpandNode(node, macros.ExpandOptions{
		LookupEnv:   opts.Lookup,
		EnvPrefix:   opts.Prefix,
		AllowedKeys: opts.AllowedKeys,
		OnExpand:    opts.OnExpand,
	})
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
