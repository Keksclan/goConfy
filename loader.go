package goconfy

import (
	"errors"
	"fmt"
	"os"
	"reflect"

	"github.com/keksclan/goConfy/explain"
	"github.com/keksclan/goConfy/internal/decode"
	"github.com/keksclan/goConfy/internal/dotenv"
	"github.com/keksclan/goConfy/internal/macros"
	"github.com/keksclan/goConfy/internal/profiles"
	"github.com/keksclan/goConfy/internal/redact"
	"github.com/keksclan/goConfy/internal/yamlparse"
	"gopkg.in/yaml.v3"
)

func collectBaseEntries[T any](node *yaml.Node, report *explain.Report, path string, opts redact.Options) {
	if node == nil || report == nil {
		return
	}

	if path == "profiles" {
		return
	}

	switch node.Kind {
	case yaml.DocumentNode:
		for _, child := range node.Content {
			collectBaseEntries[T](child, report, path, opts)
		}
	case yaml.MappingNode:
		for i := 0; i < len(node.Content)-1; i += 2 {
			key := node.Content[i].Value
			if key == "profiles" {
				continue
			}
			newPath := key
			if path != "" {
				newPath = path + "." + key
			}
			collectBaseEntries[T](node.Content[i+1], report, newPath, opts)
		}
	case yaml.SequenceNode:
		for i, child := range node.Content {
			newPath := fmt.Sprintf("%s[%d]", path, i)
			collectBaseEntries[T](child, report, newPath, opts)
		}
	case yaml.ScalarNode:
		isSecret := redact.IsSecret(reflect.TypeFor[T](), redact.StripIndices(path), opts)
		report.AddEntry(path, explain.SourceBase, node.Value, isSecret, "")
	}
}

func collectProfileEntries[T any](root *yaml.Node, report *explain.Report, profileName string, opts redact.Options) {
	if root == nil || report == nil || profileName == "" {
		return
	}

	mapping := root
	if root.Kind == yaml.DocumentNode && len(root.Content) > 0 {
		mapping = root.Content[0]
	}

	if mapping.Kind != yaml.MappingNode {
		return
	}

	var profilesNode *yaml.Node
	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == "profiles" {
			profilesNode = mapping.Content[i+1]
			break
		}
	}

	if profilesNode == nil || profilesNode.Kind != yaml.MappingNode {
		return
	}

	var overrideNode *yaml.Node
	for i := 0; i < len(profilesNode.Content)-1; i += 2 {
		if profilesNode.Content[i].Value == profileName {
			overrideNode = profilesNode.Content[i+1]
			break
		}
	}

	if overrideNode != nil && overrideNode.Kind == yaml.MappingNode {
		collectOverrides[T](overrideNode, report, "", profileName, opts)
	}
}

func collectOverrides[T any](node *yaml.Node, report *explain.Report, path string, profileName string, opts redact.Options) {
	if node == nil || report == nil {
		return
	}

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content)-1; i += 2 {
			key := node.Content[i].Value
			valNode := node.Content[i+1]
			newPath := key
			if path != "" {
				newPath = path + "." + key
			}
			collectOverrides[T](valNode, report, newPath, profileName, opts)
		}
	case yaml.SequenceNode:
		for i, child := range node.Content {
			newPath := fmt.Sprintf("%s[%d]", path, i)
			collectOverrides[T](child, report, newPath, profileName, opts)
		}
	case yaml.ScalarNode:
		isSecret := redact.IsSecret(reflect.TypeFor[T](), redact.StripIndices(path), opts)
		report.AddEntry(path, explain.SourceProfile, node.Value, isSecret, fmt.Sprintf("from profile %q", profileName))
	}
}

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

	// 2.5 Prepare Explain Report (must be before expansion to capture origins)
	var report *explain.Report
	redactOpts := redact.Options{
		Paths:        cfg.redactionPaths,
		ByConvention: cfg.redactByConvention,
	}

	if cfg.explainReporter != nil {
		report = explain.NewReport()
		collectBaseEntries[T](node, report, "", redactOpts)
	}

	// 3. Build env lookup chain (base + dotenv)
	lookup, err := buildLookupChain(cfg)
	if err != nil {
		return zero, &FieldError{Layer: "dotenv", Message: err.Error()}
	}

	// 4. Expand env macros
	expandOpts := macros.ExpandOptions{
		LookupEnv:   lookup,
		EnvPrefix:   cfg.envPrefix,
		AllowedKeys: cfg.allowedEnvKeys,
	}

	if report != nil {
		expandOpts.OnExpand = func(path, key, value string, source string) {
			isSecret := redact.IsSecret(reflect.TypeFor[T](), redact.StripIndices(path), redactOpts)
			lookupKey := key
			if source == "env" && cfg.envPrefix != "" {
				lookupKey = cfg.envPrefix + key
			}

			origin := explain.SourceEnv
			notes := ""
			switch source {
			case "env":
				origin = explain.SourceEnv
				notes = fmt.Sprintf("expanded {ENV:%s}", lookupKey)
			case "file":
				origin = explain.SourceFile
				notes = fmt.Sprintf("expanded {FILE:%s}", key)
			case "default":
				origin = explain.SourceDefault
				notes = "expanded macro default"
			}

			report.AddEntry(path, origin, value, isSecret, notes)
		}
	}

	if err := macros.ExpandNode(node, expandOpts); err != nil {
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
		if report != nil && profileName != "" {
			collectProfileEntries[T](node, report, profileName, redactOpts)
		}
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
			GetColumn() int
		}); ok {
			return zero, &FieldError{
				Layer:   "base",
				Field:   ife.GetField(),
				Line:    ife.GetLine(),
				Column:  ife.GetColumn(),
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

	if report != nil {
		cfg.explainReporter(*report)
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
