package goconfy

import (
	"errors"
	"fmt"
	"os"

	"github.com/keksclan/goConfy/internal/decode"
	"github.com/keksclan/goConfy/internal/dotenv"
	"github.com/keksclan/goConfy/internal/envmacro"
	"github.com/keksclan/goConfy/internal/profiles"
	"github.com/keksclan/goConfy/internal/yamlparse"
)

// Load reads, parses, expands, decodes, normalizes, and validates a YAML
// configuration into a typed struct T.
func Load[T any](opts ...Option) (T, error) {
	var zero T

	cfg := defaultConfig()
	for _, opt := range opts {
		opt(cfg)
	}

	// 1. Read YAML bytes
	data, err := readInput(cfg)
	if err != nil {
		return zero, &FieldError{Layer: "base", Message: err.Error()}
	}

	// 2. Parse into yaml.Node
	node, err := yamlparse.ParseBytes(data)
	if err != nil {
		return zero, &FieldError{Layer: "base", Message: err.Error()}
	}

	// 3. Build env lookup chain (base + dotenv)
	lookup, err := buildLookupChain(cfg)
	if err != nil {
		return zero, &FieldError{Layer: "dotenv", Message: err.Error()}
	}

	// 4. Expand env macros
	expandOpts := envmacro.ExpandOptions{
		Lookup:      lookup,
		Prefix:      cfg.envPrefix,
		AllowedKeys: cfg.allowedEnvKeys,
	}
	if err := envmacro.ExpandNode(node, expandOpts); err != nil {
		var fe *FieldError
		if errors.As(err, &fe) {
			fe.Layer = "base"
			return zero, fe
		}
		// Fallback for internal envmacro errors that might not be FieldError yet
		if ife, ok := err.(interface {
			GetPath() string
			GetLine() int
			GetColumn() int
		}); ok {
			return zero, &FieldError{
				Layer:   "base",
				Path:    ife.GetPath(),
				Line:    ife.GetLine(),
				Column:  ife.GetColumn(),
				Message: err.Error(),
			}
		}
		return zero, fmt.Errorf("goconfy: env macro expansion: %w", err)
	}

	// 5. Apply profile override (if enabled)
	if cfg.enableProfiles {
		profileName := profiles.SelectProfile(cfg.profile, cfg.profileEnvVar)
		if err := profiles.ApplyProfile(node, profileName); err != nil {
			return zero, &FieldError{Layer: "profile", Message: err.Error()}
		}
	}

	// 6. Marshal node back to bytes
	yamlBytes, err := yamlparse.MarshalNode(node)
	if err != nil {
		return zero, fmt.Errorf("goconfy: %w", err)
	}

	// 7. Strict decode into typed struct
	var result T
	decodeErr := func() error {
		if cfg.strictYAML {
			return decode.Strict(yamlBytes, &result)
		}
		return decode.Relaxed(yamlBytes, &result)
	}()

	if decodeErr != nil {
		if ife, ok := decodeErr.(interface {
			GetField() string
			GetLine() int
		}); ok {
			return zero, &FieldError{
				Layer:   "base",
				Field:   ife.GetField(),
				Line:    ife.GetLine(),
				Message: decodeErr.Error(),
			}
		}
		return zero, fmt.Errorf("goconfy: %w", decodeErr)
	}

	// 8. Call Normalize() if implemented
	if n, ok := any(&result).(Normalizable); ok {
		n.Normalize()
	}

	// 9. Call Validate() if implemented
	if v, ok := any(&result).(Validatable); ok {
		if err := v.Validate(); err != nil {
			return zero, fmt.Errorf("goconfy: validation: %w", err)
		}
	}

	return result, nil
}

// MustLoad is like Load but panics on error.
func MustLoad[T any](opts ...Option) T {
	cfg, err := Load[T](opts...)
	if err != nil {
		panic(err)
	}
	return cfg
}

// buildLookupChain assembles the env lookup function from base lookup and optional dotenv store.
func buildLookupChain(cfg *config) (func(string) (string, bool), error) {
	// Base lookup: user-provided or os.LookupEnv.
	base := cfg.envLookup
	if base == nil {
		base = os.LookupEnv
	}

	if !cfg.dotEnvEnabled {
		return base, nil
	}

	// Determine dotenv file path.
	path := cfg.dotEnvFile
	if path == "" {
		path = ".env"
	}

	vars, err := dotenv.ParseFile(path)
	if err != nil {
		if cfg.dotEnvOptional && errors.Is(err, os.ErrNotExist) {
			// Missing file is acceptable; use base lookup only.
			return base, nil
		}
		return nil, fmt.Errorf("dotenv: %w", err)
	}

	store := dotenv.NewStore(vars)

	// Chain lookups based on precedence.
	if cfg.dotEnvOSPrecedence {
		// OS/base wins; fallback to dotenv.
		return dotenv.ChainLookup(base, store.Lookup), nil
	}
	// Dotenv wins; fallback to OS/base.
	return dotenv.ChainLookup(store.Lookup, base), nil
}

func readInput(cfg *config) ([]byte, error) {
	if cfg.bytes != nil {
		return cfg.bytes, nil
	}
	if cfg.file != "" {
		data, err := os.ReadFile(cfg.file)
		if err != nil {
			return nil, fmt.Errorf("read file %q: %w", cfg.file, err)
		}
		return data, nil
	}
	return nil, fmt.Errorf("no input source: use WithFile or WithBytes")
}
