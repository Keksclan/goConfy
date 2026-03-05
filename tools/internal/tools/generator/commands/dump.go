package commands

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/keksclan/goConfy/tools/generator/registry"
	"github.com/keksclan/goConfy/internal/decode"
	"github.com/keksclan/goConfy/internal/dotenv"
	"github.com/keksclan/goConfy/internal/envmacro"
	"github.com/keksclan/goConfy/internal/profiles"
	"github.com/keksclan/goConfy/internal/redact"
	"github.com/keksclan/goConfy/internal/yamlparse"
)

// RunDump implements the "dump" subcommand. It loads a YAML config file,
// decodes it into the registered config type, and prints a redacted JSON
// representation to stdout. Fields tagged with secret:"true" are replaced
// with "[REDACTED]". Accepts -id, -in, -dotenv, and -profile flags.
func RunDump(args []string) error {
	fs := flag.NewFlagSet("dump", flag.ContinueOnError)

	id := fs.String("id", "", "registry ID of the config type (required)")
	in := fs.String("in", "", "input YAML file path (required)")
	dotenvFile := fs.String("dotenv", "", "optional dotenv file path")
	profile := fs.String("profile", "", "profile override")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *id == "" {
		return fmt.Errorf("-id is required; registered IDs: %v", registry.List())
	}
	if *in == "" {
		return fmt.Errorf("-in is required")
	}

	provider, ok := registry.Get(*id)
	if !ok {
		return fmt.Errorf("no provider registered with id %q", *id)
	}

	inPath, err := resolveInputPath(*in)
	if err != nil {
		return err
	}

	data, err := os.ReadFile(inPath)
	if err != nil {
		return fmt.Errorf("read %q: %w", inPath, err)
	}

	node, err := yamlparse.ParseBytes(data)
	if err != nil {
		return fmt.Errorf("parse YAML %q: %w", inPath, err)
	}

	// Build lookup chain.
	lookup := os.LookupEnv
	if *dotenvFile != "" {
		vars, err := dotenv.ParseFile(*dotenvFile)
		if err != nil {
			return fmt.Errorf("parse dotenv %q: %w", *dotenvFile, err)
		}
		store := dotenv.NewStore(vars)
		lookup = dotenv.ChainLookup(lookup, store.Lookup)
	}

	// Expand env macros.
	if err := envmacro.ExpandNode(node, envmacro.ExpandOptions{Lookup: lookup}); err != nil {
		return fmt.Errorf("env macro expansion: %w", err)
	}

	// Apply profile.
	if *profile != "" {
		if err := profiles.ApplyProfile(node, *profile); err != nil {
			return fmt.Errorf("profile apply: %w", err)
		}
	}

	// Marshal back to bytes.
	yamlBytes, err := yamlparse.MarshalNode(node)
	if err != nil {
		return fmt.Errorf("marshal YAML: %w", err)
	}

	// Decode into typed struct.
	result := provider.New()
	if err := decode.Relaxed(yamlBytes, result); err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	// Redact and dump as JSON.
	redacted := redact.Apply(result, redact.Options{})
	out, err := json.MarshalIndent(redacted, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal redacted JSON: %w", err)
	}

	fmt.Println(string(out))
	return nil
}
