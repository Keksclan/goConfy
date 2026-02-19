// Package logic provides helper functions that wrap the core goconfy pipeline
// for use by TUI screens. Each function accepts simple parameters (file paths,
// flags) and returns ready-to-display results or errors.
package logic

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/keksclan/goConfy/gen/registry"
	"github.com/keksclan/goConfy/gen/yamltemplate"
	"github.com/keksclan/goConfy/internal/decode"
	"github.com/keksclan/goConfy/internal/dotenv"
	"github.com/keksclan/goConfy/internal/envmacro"
	"github.com/keksclan/goConfy/internal/profiles"
	"github.com/keksclan/goConfy/internal/redact"
	"github.com/keksclan/goConfy/internal/tui/state"
	"github.com/keksclan/goConfy/internal/yamlparse"
	"gopkg.in/yaml.v3"
)

// LoadRawYAML reads a YAML file and returns its contents as a string.
func LoadRawYAML(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read %q: %w", path, err)
	}
	return string(data), nil
}

// buildLookup assembles the env lookup chain from state config.
func buildLookup(cfg *state.Config) (func(string) (string, bool), error) {
	lookup := os.LookupEnv
	if cfg.DotEnvPath != "" {
		vars, err := dotenv.ParseFile(cfg.DotEnvPath)
		if err != nil {
			if cfg.DotEnvOptional && os.IsNotExist(err) {
				return lookup, nil
			}
			return nil, fmt.Errorf("parse dotenv %q: %w", cfg.DotEnvPath, err)
		}
		store := dotenv.NewStore(vars)
		if cfg.DotEnvOSWins {
			lookup = dotenv.ChainLookup(lookup, store.Lookup)
		} else {
			lookup = dotenv.ChainLookup(store.Lookup, lookup)
		}
	}
	return lookup, nil
}

// parseAndExpand reads, parses, and expands macros in a YAML file.
// Returns the yaml.Node after expansion (before profile merge).
func parseAndExpand(cfg *state.Config) (*yaml.Node, error) {
	data, err := os.ReadFile(cfg.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", cfg.ConfigPath, err)
	}
	node, err := yamlparse.ParseBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}
	lookup, err := buildLookup(cfg)
	if err != nil {
		return nil, err
	}
	if err := envmacro.ExpandNode(node, envmacro.ExpandOptions{Lookup: lookup}); err != nil {
		return nil, fmt.Errorf("env macro expansion: %w", err)
	}
	return node, nil
}

// ExpandedYAML returns the YAML content after macro expansion (before profile merge).
func ExpandedYAML(cfg *state.Config) (string, error) {
	node, err := parseAndExpand(cfg)
	if err != nil {
		return "", err
	}
	b, err := yamlparse.MarshalNode(node)
	if err != nil {
		return "", fmt.Errorf("marshal YAML: %w", err)
	}
	return string(b), nil
}

// MergedYAML returns the YAML content after macro expansion and profile merge.
func MergedYAML(cfg *state.Config) (string, error) {
	node, err := parseAndExpand(cfg)
	if err != nil {
		return "", err
	}
	profile := profiles.SelectProfile(cfg.ActiveProfile, cfg.ProfileEnvVar)
	if err := profiles.ApplyProfile(node, profile); err != nil {
		return "", fmt.Errorf("profile apply: %w", err)
	}
	b, err := yamlparse.MarshalNode(node)
	if err != nil {
		return "", fmt.Errorf("marshal YAML: %w", err)
	}
	return string(b), nil
}

// ValidateConfig loads and validates using a registry provider. Returns nil on
// success, or the validation error.
func ValidateConfig(cfg *state.Config, providerID string) error {
	provider, ok := registry.Get(providerID)
	if !ok {
		return fmt.Errorf("no provider registered with id %q; registered: %v", providerID, registry.List())
	}
	node, err := parseAndExpand(cfg)
	if err != nil {
		return err
	}
	profile := profiles.SelectProfile(cfg.ActiveProfile, cfg.ProfileEnvVar)
	if err := profiles.ApplyProfile(node, profile); err != nil {
		return fmt.Errorf("profile apply: %w", err)
	}
	yamlBytes, err := yamlparse.MarshalNode(node)
	if err != nil {
		return fmt.Errorf("marshal YAML: %w", err)
	}
	result := provider.New()
	if cfg.StrictMode {
		if err := decode.Strict(yamlBytes, result); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	} else {
		if err := decode.Relaxed(yamlBytes, result); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}
	return nil
}

// DumpRedactedJSON loads the config via a registry provider and returns a
// redacted JSON string. Secrets tagged with secret:"true" are replaced.
func DumpRedactedJSON(cfg *state.Config, providerID string) (string, error) {
	provider, ok := registry.Get(providerID)
	if !ok {
		return "", fmt.Errorf("no provider registered with id %q", providerID)
	}
	node, err := parseAndExpand(cfg)
	if err != nil {
		return "", err
	}
	profile := profiles.SelectProfile(cfg.ActiveProfile, cfg.ProfileEnvVar)
	if err := profiles.ApplyProfile(node, profile); err != nil {
		return "", fmt.Errorf("profile apply: %w", err)
	}
	yamlBytes, err := yamlparse.MarshalNode(node)
	if err != nil {
		return "", fmt.Errorf("marshal YAML: %w", err)
	}
	result := provider.New()
	if err := decode.Relaxed(yamlBytes, result); err != nil {
		return "", fmt.Errorf("decode: %w", err)
	}
	redacted := redact.Apply(result, redact.Options{Paths: cfg.RedactionPaths})
	out, err := json.MarshalIndent(redacted, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal redacted JSON: %w", err)
	}
	return string(out), nil
}

// FormatYAML reads, optionally expands macros, and re-marshals the YAML.
// Returns the formatted bytes.
func FormatYAML(cfg *state.Config) ([]byte, error) {
	data, err := os.ReadFile(cfg.ConfigPath)
	if err != nil {
		return nil, fmt.Errorf("read %q: %w", cfg.ConfigPath, err)
	}
	node, err := yamlparse.ParseBytes(data)
	if err != nil {
		return nil, fmt.Errorf("parse YAML: %w", err)
	}
	if cfg.ExpandMacros {
		lookup, err := buildLookup(cfg)
		if err != nil {
			return nil, err
		}
		if err := envmacro.ExpandNode(node, envmacro.ExpandOptions{Lookup: lookup}); err != nil {
			return nil, fmt.Errorf("env macro expansion: %w", err)
		}
	}
	if !cfg.KeepProfiles {
		profile := profiles.SelectProfile(cfg.ActiveProfile, cfg.ProfileEnvVar)
		if err := profiles.ApplyProfile(node, profile); err != nil {
			return nil, fmt.Errorf("profile apply: %w", err)
		}
	}
	return yamlparse.MarshalNode(node)
}

// WriteFormatted writes formatted YAML bytes to the resolved output path.
func WriteFormatted(cfg *state.Config, data []byte) (string, error) {
	outPath := cfg.ConfigPath
	if cfg.OutputPath != "" {
		outPath = cfg.OutputPath
	}
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return "", fmt.Errorf("create directories: %w", err)
	}
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		return "", fmt.Errorf("write %q: %w", outPath, err)
	}
	return outPath, nil
}

// GenerateTemplate generates a YAML template from a registry provider.
// Returns the generated YAML string and optional .env content.
func GenerateTemplate(cfg *state.Config) (yamlContent string, envContent string, err error) {
	provider, ok := registry.Get(cfg.RegistryID)
	if !ok {
		return "", "", fmt.Errorf("no provider registered with id %q; registered: %v", cfg.RegistryID, registry.List())
	}
	inst := provider.New()
	opts := yamltemplate.Options{}
	if cfg.IncludeProfiles {
		opts.Profile = "dev,staging,prod"
	}
	opts.DotEnv = cfg.GenerateDotEnv

	yamlContent, err = yamltemplate.Generate(inst, opts)
	if err != nil {
		return "", "", fmt.Errorf("generate YAML template: %w", err)
	}

	if cfg.GenerateDotEnv {
		envContent, err = yamltemplate.GenerateDotEnv(inst)
		if err != nil {
			return yamlContent, "", fmt.Errorf("generate .env: %w", err)
		}
	}
	return yamlContent, envContent, nil
}

// WriteTemplate writes the generated template files to disk.
func WriteTemplate(cfg *state.Config, yamlContent, envContent string) ([]string, error) {
	outPath := cfg.OutputPath
	if outPath == "" {
		outPath = "./config.yml"
	}
	// If output is a directory, append default filename.
	info, err := os.Stat(outPath)
	if err == nil && info.IsDir() {
		outPath = filepath.Join(outPath, "config.yml")
	}

	if !cfg.ForceOverwrite {
		if _, err := os.Stat(outPath); err == nil {
			return nil, fmt.Errorf("file %q already exists; enable force overwrite", outPath)
		}
	}

	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return nil, fmt.Errorf("create directories: %w", err)
	}

	var written []string
	if err := os.WriteFile(outPath, []byte(yamlContent), 0o644); err != nil {
		return nil, fmt.Errorf("write %q: %w", outPath, err)
	}
	written = append(written, outPath)

	if envContent != "" {
		envPath := filepath.Join(filepath.Dir(outPath), ".env")
		if !cfg.ForceOverwrite {
			if _, err := os.Stat(envPath); err == nil {
				return written, fmt.Errorf("file %q already exists; enable force overwrite", envPath)
			}
		}
		if err := os.WriteFile(envPath, []byte(envContent), 0o644); err != nil {
			return written, fmt.Errorf("write %q: %w", envPath, err)
		}
		written = append(written, envPath)
	}

	return written, nil
}

// ListProfiles parses a YAML file and returns the list of profile names found
// under the "profiles" top-level key, plus whether the key exists at all.
func ListProfiles(configPath string) ([]string, bool, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, false, fmt.Errorf("read %q: %w", configPath, err)
	}
	node, err := yamlparse.ParseBytes(data)
	if err != nil {
		return nil, false, fmt.Errorf("parse YAML: %w", err)
	}

	mapping := node
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		mapping = node.Content[0]
	}
	if mapping.Kind != yaml.MappingNode {
		return nil, false, nil
	}

	for i := 0; i < len(mapping.Content)-1; i += 2 {
		if mapping.Content[i].Value == "profiles" {
			profilesNode := mapping.Content[i+1]
			if profilesNode.Kind != yaml.MappingNode {
				return nil, true, nil
			}
			var names []string
			for j := 0; j < len(profilesNode.Content)-1; j += 2 {
				names = append(names, profilesNode.Content[j].Value)
			}
			return names, true, nil
		}
	}
	return nil, false, nil
}

// ListRegistryIDs returns all registered provider IDs.
func ListRegistryIDs() []string {
	return registry.List()
}
