// Package yamltemplate generates YAML config templates from Go struct types
// using reflection and struct tags.
//
// It inspects struct fields for tags such as yaml, env, default, desc, example,
// required, secret, and sep, then emits a well-formatted YAML file with:
//   - Environment macro placeholders: {ENV:KEY:DEFAULT}
//   - Descriptive comments above each field
//   - Default values and type-appropriate zero values
//   - Secret redaction (never outputs real secret values)
//   - Optional profile override skeletons
//   - Optional .env file generation
//
// Primary entry points are [Generate] and [GenerateDotEnv].
package yamltemplate
