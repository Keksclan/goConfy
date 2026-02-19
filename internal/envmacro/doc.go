// Package envmacro implements exact-match environment variable macro expansion
// for YAML configuration values.
//
// Macros use the format {ENV:KEY} or {ENV:KEY:default} and are only expanded
// when the entire YAML scalar value matches the pattern. This exact-match
// design is a deliberate security choice: it prevents accidental expansion of
// partial strings, shell injection, and recursive variable references.
//
// The expansion is controlled by [ExpandOptions], which supports:
//   - Custom lookup functions (for dotenv integration)
//   - Environment key prefix filtering
//   - Explicit allowlists of permitted keys
//
// This is an internal package used by the goconfy loader pipeline.
package envmacro
