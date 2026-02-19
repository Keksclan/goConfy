package envmacro

import "regexp"

// EnvMacroRegex is the strict regex for environment macros.
// Pattern: {ENV:KEY} or {ENV:KEY:default}
var EnvMacroRegex = regexp.MustCompile(`^\{ENV:([A-Z0-9_]+)(?::([^}]*))?\}$`)
