// Package yamlparse provides low-level YAML parsing and marshaling helpers
// that operate on [yaml.Node] trees.
//
// Working with yaml.Node (rather than decoding directly into structs) allows
// the goconfy pipeline to manipulate the YAML tree—expanding macros, merging
// profiles—before the final typed decode step.
//
// This is an internal package used by the goconfy loader pipeline.
package yamlparse
