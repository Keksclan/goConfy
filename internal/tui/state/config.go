// Package state holds the shared TUI state: selected paths, toggles, chosen
// profile, and user settings. Every screen reads from and writes to this
// central struct so that data flows through the application without globals.
package state

// Config holds all mutable TUI state shared across screens.
type Config struct {
	// File paths
	ConfigPath string // YAML config path (required for most ops)
	DotEnvPath string // optional .env path
	OutputPath string // optional output path for fmt / init

	// Profile
	ProfileEnvVar string // env var name, default APP_PROFILE
	ActiveProfile string // explicitly chosen profile override

	// Toggles
	StrictMode      bool // strict YAML decoding (default true)
	DotEnvOptional  bool // missing .env is not an error
	DotEnvOSWins    bool // OS env takes precedence over .env
	ExpandMacros    bool // expand macros before formatting
	KeepProfiles    bool // keep profiles section in fmt output
	GenerateDotEnv  bool // generate .env alongside init output
	IncludeProfiles bool // include profiles skeleton in init output
	ForceOverwrite  bool // overwrite existing files without asking

	// Init screen
	RegistryID string // selected registry provider ID

	// Redaction
	RedactionPaths []string // dot-paths to redact in YAML preview
}

// DefaultConfig returns a Config with sensible defaults matching the library
// defaults (strict on, OS env wins, APP_PROFILE).
func DefaultConfig() *Config {
	return &Config{
		StrictMode:     true,
		DotEnvOptional: false,
		DotEnvOSWins:   true,
		ProfileEnvVar:  "APP_PROFILE",
		KeepProfiles:   true,
		RedactionPaths: []string{
			"redis.password",
			"auth.opaque.client_secret",
			"postgres.url",
		},
	}
}
