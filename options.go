package goconfy

import (
	"github.com/keksclan/goConfy/explain"
)

// Option configures the Load pipeline.
type Option func(*config)

// RedactOption configures the redaction behavior.
type RedactOption func(*redactConfig)

// Normalizable can be implemented by config structs to automatically normalize values
// after the YAML decoding phase but before validation.
// This is useful for trimming strings, setting defaults for missing fields, or converting values.
type Normalizable interface {
	// Normalize is called by [Load] after successful decoding.
	Normalize()
}

// Validatable can be implemented by config structs to automatically validate values
// after the decoding and normalization phases.
// If Validate returns an error, [Load] will fail and return that error.
type Validatable interface {
	// Validate is called by [Load] after decoding and normalization.
	Validate() error
}

type config struct {
	file                string
	bytes               []byte
	envLookup           func(string) (string, bool)
	envPrefix           string
	allowedEnvKeys      []string
	strictYAML          bool
	enableProfiles      bool
	profileEnvVar       string
	profile             string
	redactionPaths      []string
	redactByConvention  bool
	inlineInterpolation bool

	explainReporter explain.Reporter

	// Dotenv options
	dotEnvEnabled      bool
	dotEnvFile         string
	dotEnvOSPrecedence bool
	dotEnvOptional     bool
}

type redactConfig struct {
	paths        []string
	byConvention bool
}

func defaultConfig() *config {
	return &config{
		strictYAML:         true,
		profileEnvVar:      "APP_PROFILE",
		dotEnvOSPrecedence: true,
	}
}

// WithFile sets the file path from which the YAML configuration will be read.
// If the file is missing or unreadable, [Load] will return an error.
func WithFile(path string) Option {
	return func(c *config) {
		c.file = path
	}
}

// WithBytes provides the YAML configuration directly as a byte slice.
// This is useful for tests or when embedding configuration files.
func WithBytes(data []byte) Option {
	return func(c *config) {
		c.bytes = data
	}
}

// WithEnvLookup sets a custom function to look up environment variables during macro expansion.
// By default, [os.LookupEnv] is used. This can be used to provide a mock environment for tests.
func WithEnvLookup(fn func(string) (string, bool)) Option {
	return func(c *config) {
		c.envLookup = fn
	}
}

// WithEnvPrefix sets a prefix that will be automatically prepended to all environment keys
// during macro expansion.
// For example, WithEnvPrefix("MYAPP_") will make {ENV:PORT:8080} look up "MYAPP_PORT".
func WithEnvPrefix(prefix string) Option {
	return func(c *config) {
		c.envPrefix = prefix
	}
}

// WithAllowedEnvKeys restricts environment macro expansion to the specified keys.
// If this list is non-empty, any {ENV:KEY} macro using a KEY not in this list will return an error.
// This provides an extra layer of security by preventing accidental exposure of sensitive environment variables.
func WithAllowedEnvKeys(keys []string) Option {
	return func(c *config) {
		c.allowedEnvKeys = keys
	}
}

// WithStrictYAML enables or disables strict YAML decoding.
// When enabled (default: true), [Load] will return an error if the YAML contains keys
// that are not defined in the target configuration struct. This is recommended
// to catch typos in configuration files.
func WithStrictYAML(strict bool) Option {
	return func(c *config) {
		c.strictYAML = strict
	}
}

// WithEnableProfiles enables or disables profile-based overrides.
// When enabled (default: false), goConfy will look for a "profiles" key in the root of the YAML
// and merge the selected profile's values into the base configuration.
//
// Profile merging is a deep merge: sub-mappings are merged, while scalars and sequences
// from the profile override those in the base configuration. The "profiles" key itself
// is removed from the YAML tree before decoding into the target struct.
func WithEnableProfiles(enabled bool) Option {
	return func(c *config) {
		c.enableProfiles = enabled
	}
}

// WithProfileEnvVar sets the name of the environment variable used to select the active profile.
// Default is "APP_PROFILE". This is only used if [WithEnableProfiles] is true.
func WithProfileEnvVar(envVar string) Option {
	return func(c *config) {
		c.profileEnvVar = envVar
	}
}

// WithProfile explicitly sets the active profile name, bypassing the environment variable lookup.
// This is useful for programmatic profile selection (e.g., via a CLI flag).
func WithProfile(profile string) Option {
	return func(c *config) {
		c.profile = profile
	}
}

// WithRedactionPaths sets dot-separated paths to redact in output.
func WithRedactionPaths(paths []string) Option {
	return func(c *config) {
		c.redactionPaths = paths
	}
}

// WithInlineInterpolation enables or disables inline interpolation (default: false).
// This is reserved for future extension.
func WithInlineInterpolation(enabled bool) Option {
	return func(c *config) {
		c.inlineInterpolation = enabled
	}
}

// WithDotEnvEnabled enables or disables the loading of .env files.
// When enabled, [Load] will look for a .env file (default path: ".env") and add its
// contents to the environment lookup chain for macro expansion.
//
// IMPORTANT: This does NOT mutate the OS environment (os.Environ). The values
// from the .env file are only available to the [Load] pipeline and its macros.
// Use [WithDotEnvOSPrecedence] to control whether the OS environment or the .env file wins.
func WithDotEnvEnabled(enabled bool) Option {
	return func(c *config) {
		c.dotEnvEnabled = enabled
	}
}

// WithDotEnvFile sets the path to the .env file and implicitly enables dotenv loading.
// Default path is ".env".
func WithDotEnvFile(path string) Option {
	return func(c *config) {
		c.dotEnvFile = path
		c.dotEnvEnabled = true
	}
}

// WithDotEnvOSPrecedence controls whether variables from the OS environment take precedence
// over variables from the .env file.
// Default is true (OS environment wins). If set to false, the .env file wins.
func WithDotEnvOSPrecedence(osPrecedence bool) Option {
	return func(c *config) {
		c.dotEnvOSPrecedence = osPrecedence
	}
}

// WithDotEnvOptional controls whether a missing .env file should be treated as an error.
// When true, [Load] will silently ignore a missing .env file.
// Default is false (missing file is an error).
func WithDotEnvOptional(optional bool) Option {
	return func(c *config) {
		c.dotEnvOptional = optional
	}
}

// WithRedactByConvention enables automatic redaction based on field names
// (e.g., password, secret, token, key, private).
func WithRedactByConvention(enabled bool) Option {
	return func(c *config) {
		c.redactByConvention = enabled
	}
}

// WithRedactPaths sets dot-separated paths to redact.
func WithRedactPaths(paths []string) RedactOption {
	return func(c *redactConfig) {
		c.paths = paths
	}
}

// WithRedactByConventionOption enables automatic redaction based on field names for a specific Redacted call.
func WithRedactByConventionOption(enabled bool) RedactOption {
	return func(c *redactConfig) {
		c.byConvention = enabled
	}
}

// WithExplainReporter enables configuration reporting and sets the reporter function.
func WithExplainReporter(reporter explain.Reporter) Option {
	return func(c *config) {
		c.explainReporter = reporter
	}
}
