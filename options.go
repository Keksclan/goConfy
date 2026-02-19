package goconfy

// Option configures the Load pipeline.
type Option func(*config)

// RedactOption configures the redaction behavior.
type RedactOption func(*redactConfig)

// Normalizable can be implemented by config structs to normalize values after decoding.
type Normalizable interface {
	Normalize()
}

// Validatable can be implemented by config structs to validate values after decoding.
type Validatable interface {
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
	inlineInterpolation bool

	// Dotenv options
	dotEnvEnabled      bool
	dotEnvFile         string
	dotEnvOSPrecedence bool
	dotEnvOptional     bool
}

type redactConfig struct {
	paths []string
}

func defaultConfig() *config {
	return &config{
		strictYAML:         true,
		profileEnvVar:      "APP_PROFILE",
		dotEnvOSPrecedence: true,
	}
}

// WithFile sets the YAML config file path.
func WithFile(path string) Option {
	return func(c *config) {
		c.file = path
	}
}

// WithBytes sets the YAML config bytes directly.
func WithBytes(data []byte) Option {
	return func(c *config) {
		c.bytes = data
	}
}

// WithEnvLookup sets a custom environment variable lookup function.
func WithEnvLookup(fn func(string) (string, bool)) Option {
	return func(c *config) {
		c.envLookup = fn
	}
}

// WithEnvPrefix sets a prefix that will be prepended to environment variable keys.
func WithEnvPrefix(prefix string) Option {
	return func(c *config) {
		c.envPrefix = prefix
	}
}

// WithAllowedEnvKeys restricts which environment keys may be expanded.
func WithAllowedEnvKeys(keys []string) Option {
	return func(c *config) {
		c.allowedEnvKeys = keys
	}
}

// WithStrictYAML enables or disables strict YAML decoding (default: true).
func WithStrictYAML(strict bool) Option {
	return func(c *config) {
		c.strictYAML = strict
	}
}

// WithEnableProfiles enables or disables profile-based overrides.
func WithEnableProfiles(enabled bool) Option {
	return func(c *config) {
		c.enableProfiles = enabled
	}
}

// WithProfileEnvVar sets the environment variable name used to select a profile (default: APP_PROFILE).
func WithProfileEnvVar(envVar string) Option {
	return func(c *config) {
		c.profileEnvVar = envVar
	}
}

// WithProfile explicitly sets the active profile name.
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

// WithDotEnvEnabled enables or disables .env file loading (default: false).
func WithDotEnvEnabled(enabled bool) Option {
	return func(c *config) {
		c.dotEnvEnabled = enabled
	}
}

// WithDotEnvFile sets the path to the .env file.
// Setting this implies dotenv is enabled.
func WithDotEnvFile(path string) Option {
	return func(c *config) {
		c.dotEnvFile = path
		c.dotEnvEnabled = true
	}
}

// WithDotEnvOSPrecedence controls whether OS/base env takes precedence over .env values (default: true).
// When true, OS env wins; when false, .env wins.
func WithDotEnvOSPrecedence(osPrecedence bool) Option {
	return func(c *config) {
		c.dotEnvOSPrecedence = osPrecedence
	}
}

// WithDotEnvOptional controls whether a missing .env file is an error (default: false).
// When true, a missing file is silently ignored.
func WithDotEnvOptional(optional bool) Option {
	return func(c *config) {
		c.dotEnvOptional = optional
	}
}

// WithRedactPaths sets dot-separated paths to redact.
func WithRedactPaths(paths []string) RedactOption {
	return func(c *redactConfig) {
		c.paths = paths
	}
}
