// Package profiles implements profile-based configuration overrides.
//
// A YAML config file may contain a top-level "profiles" mapping with named
// override sections (e.g. "dev", "staging", "prod"). When a profile is
// selected—either explicitly or via an environment variable—the matching
// override is deep-merged into the base configuration.
//
// Merge semantics: profile values override base values at the leaf level.
// Nested mappings are merged recursively; scalars and sequences are replaced.
//
// This is an internal package used by the goconfy loader pipeline.
package profiles
