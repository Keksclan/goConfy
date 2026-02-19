// Package decode provides strict and relaxed YAML decoding into Go structs.
//
// [Strict] rejects unknown fields in the YAML input, making it suitable for
// production configs where typos should be caught early. [Relaxed] silently
// ignores unknown fields, useful for forward-compatible or partial decoding.
//
// This is an internal package used by the goconfy loader pipeline.
package decode
