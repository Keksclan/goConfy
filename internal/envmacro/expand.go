package envmacro

import (
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
