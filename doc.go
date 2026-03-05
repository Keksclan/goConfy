// Package goconfy provides a strongly typed, strict, and secure YAML configuration loader for Go.
// It is designed to be a complete configuration pipeline, handling everything from parsing
// to validation and secret redaction.
//
// # Key Features
//
//   - YAML parsing via gopkg.in/yaml.v3
//   - Exact environment macro expansion: {ENV:KEY:default}
//   - Optional profile-based overrides (e.g., dev, staging, prod)
//   - .env file loading without mutating OS environment (os.Environ)
//   - Strict decoding into typed structs (rejects unknown fields by default)
//   - Validation and normalization hooks (implemented via interfaces)
//   - Secret redaction for safe logging and JSON dumping
//
// # Configuration Pipeline
//
// The [Load] function executes the following pipeline:
//
//  1. Read Input: From a file ([WithFile]) or direct bytes ([WithBytes]).
//  2. Parse: Initial YAML parsing into an intermediate tree representation (via yaml.v3).
//  3. Environment: Assemble the lookup chain (base OS environment + optional .env).
//  4. Macro Expansion: Resolve macros in the YAML tree before decoding.
//  5. Profiles: Apply overrides from the "profiles" section if enabled ([WithEnableProfiles]).
//  6. Decode: Map the expanded YAML tree into the target struct T. Strict mode is enabled by default.
//  7. Normalize: Call the optional [Normalizable] interface if implemented by T.
//  8. Validate: Call the optional [Validatable] interface if implemented by T.
//
// # Macros
//
// goConfy supports two types of macros that are expanded within YAML scalar values:
//
//   - {ENV:KEY:default}: Looks up KEY in the environment lookup chain.
//   - {FILE:PATH:default}: Reads the content of the file at PATH.
//
// Macros must match the pattern exactly. If a key or file is missing and no default is provided,
// the expansion will result in an error. Use [WithEnvPrefix] to automatically prefix environment keys.
//
// # Precedence & Dotenv
//
// When [.env] loading is enabled via [WithDotEnvEnabled] or [WithDotEnvFile], goConfy builds
// a lookup chain for environment macros. By default, the OS environment takes precedence
// over [.env] file values. This can be reversed using [WithDotEnvOSPrecedence].
//
// # Profiles
//
// Profiles allow you to define environment-specific overrides within the same YAML file.
// If enabled, the loader looks for a top-level "profiles" key. The active profile is selected
// via environment variable (default: APP_PROFILE) or [WithProfile].
//
//	# Example YAML with profiles
//	port: 8080
//	profiles:
//	  prod:
//	    port: 80
//
// In this case, if the "prod" profile is active, the final port value will be 80.
//
// # Error Handling
//
// Failures in the pipeline return a [FieldError], which contains information about the
// phase ([Layer]), the dot-path to the field ([Path]), and the line/column numbers in the
// YAML source. Multiple errors may be returned as a [MultiError].
//
// # Basic Example
//
//	type Config struct {
//	    Port    int            `yaml:"port" default:"8080"`
//	    Host    string         `yaml:"host" default:"localhost"`
//	    Timeout types.Duration `yaml:"timeout"`
//	}
//
//	func main() {
//	    cfg, err := goconfy.Load[Config](
//	        goconfy.WithFile("config.yml"),
//	        goconfy.WithDotEnvEnabled(true),
//	    )
//	    if err != nil {
//	        log.Fatalf("failed to load config: %v", err)
//	    }
//	    fmt.Printf("Server starting on %s:%d\n", cfg.Host, cfg.Port)
//	}
package goconfy
