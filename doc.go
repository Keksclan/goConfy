// Package goconfy provides a strongly typed, strict, and secure YAML configuration loader.
//
// goconfy supports:
//   - YAML parsing via yaml.v3
//   - Exact environment macro expansion: {ENV:KEY:default}
//   - Optional profile-based overrides
//   - Strict decoding into typed structs
//   - Validation and normalization hooks
//   - Secret redaction helpers
//
// Basic usage:
//
//	type Config struct {
//	    Port int    `yaml:"port"`
//	    Host string `yaml:"host"`
//	}
//
//	cfg, err := goconfy.Load[Config](goconfy.WithFile("config.yml"))
package goconfy
