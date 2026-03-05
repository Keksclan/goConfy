// Package commands implements the subcommands for the goconfygen CLI tool.
//
// Each subcommand is exposed as a public Run function:
//   - [RunInit]: generates a YAML config template from a registered config type
//   - [RunValidate]: validates an existing YAML file against a typed config struct
//   - [RunFmt]: canonicalizes YAML formatting with optional macro expansion
//   - [RunDump]: dumps a redacted JSON representation of the loaded config
//
// Subcommands use the [generator/registry] package to discover config types at
// runtime. The target project must register its config type via a [registry.Provider]
// implementation before invoking the CLI.
package commands
