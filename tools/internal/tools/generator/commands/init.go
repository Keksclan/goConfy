package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/keksclan/goConfy/tools/generator/registry"
	"github.com/keksclan/goConfy/tools/generator/yamltemplate"
)

// RunInit implements the "init" subcommand. It generates a YAML config
// template from a registered config type, emitting keys with default values,
// environment macro placeholders, and descriptive comments derived from struct
// tags. Accepts -id, -out, -profile, -dotenv, -force, and -mkdir flags.
func RunInit(args []string) error {
	fs := flag.NewFlagSet("init", flag.ContinueOnError)

	id := fs.String("id", "", "registry ID of the config type (required)")
	pkg := fs.String("pkg", "", "Go package import path (not yet supported; use -id)")
	typeName := fs.String("type", "", "Go type name (not yet supported; use -id)")
	out := fs.String("out", "./config.yml", "output file path")
	profile := fs.String("profile", "", "include profiles skeleton (e.g. dev, staging, prod)")
	dotenv := fs.Bool("dotenv", false, "also generate a sample .env file")
	force := fs.Bool("force", false, "overwrite existing files")
	mkdir := fs.Bool("mkdir", true, "create parent directories if needed")

	if err := fs.Parse(args); err != nil {
		return err
	}

	// Inform about unsupported -pkg/-type flags.
	if *pkg != "" || *typeName != "" {
		fmt.Fprintln(os.Stderr, "goconfygen init: -pkg and -type are not yet supported. Use -id with the registry approach.")
		if *id == "" {
			return fmt.Errorf("-id is required (registry approach)")
		}
	}

	if *id == "" {
		return fmt.Errorf("-id is required; registered IDs: %v", registry.List())
	}

	provider, ok := registry.Get(*id)
	if !ok {
		return fmt.Errorf("no provider registered with id %q; registered IDs: %v", *id, registry.List())
	}

	// Resolve output path.
	outPath, err := resolveOutputPath(*out, "config.yml")
	if err != nil {
		return err
	}

	// Check if file exists.
	if !*force {
		if _, err := os.Stat(outPath); err == nil {
			return fmt.Errorf("file %q already exists; use -force to overwrite", outPath)
		}
	}

	// Create parent directories.
	if *mkdir {
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("create directories for %q: %w", outPath, err)
		}
	}

	// Generate YAML template.
	cfgInstance := provider.New()
	opts := yamltemplate.Options{
		Profile: *profile,
		DotEnv:  *dotenv,
	}

	yamlContent, err := yamltemplate.Generate(cfgInstance, opts)
	if err != nil {
		return fmt.Errorf("generate YAML template: %w", err)
	}

	if err := os.WriteFile(outPath, []byte(yamlContent), 0o644); err != nil {
		return fmt.Errorf("write %q: %w", outPath, err)
	}
	fmt.Fprintf(os.Stderr, "wrote %s\n", outPath)

	// Generate .env file if requested.
	if *dotenv {
		envPath := filepath.Join(filepath.Dir(outPath), ".env")
		if !*force {
			if _, err := os.Stat(envPath); err == nil {
				return fmt.Errorf("file %q already exists; use -force to overwrite", envPath)
			}
		}

		envContent, err := yamltemplate.GenerateDotEnv(cfgInstance)
		if err != nil {
			return fmt.Errorf("generate .env: %w", err)
		}

		if err := os.WriteFile(envPath, []byte(envContent), 0o644); err != nil {
			return fmt.Errorf("write %q: %w", envPath, err)
		}
		fmt.Fprintf(os.Stderr, "wrote %s\n", envPath)
	}

	return nil
}
