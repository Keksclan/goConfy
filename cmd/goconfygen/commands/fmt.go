package commands

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/keksclan/goConfy/internal/dotenv"
	"github.com/keksclan/goConfy/internal/envmacro"
	"github.com/keksclan/goConfy/internal/yamlparse"
)

// RunFmt implements the "fmt" subcommand.
func RunFmt(args []string) error {
	fs := flag.NewFlagSet("fmt", flag.ContinueOnError)

	in := fs.String("in", "", "input YAML file path (required)")
	out := fs.String("out", "", "output file path (default: overwrite input)")
	dotenvFile := fs.String("dotenv", "", "optional dotenv file path for macro expansion")
	expand := fs.Bool("expand", false, "expand env macros before formatting")
	mkdir := fs.Bool("mkdir", true, "create parent directories if needed")

	if err := fs.Parse(args); err != nil {
		return err
	}

	if *in == "" {
		return fmt.Errorf("-in is required")
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

	// Optionally expand macros.
	if *expand {
		lookup := os.LookupEnv
		if *dotenvFile != "" {
			vars, err := dotenv.ParseFile(*dotenvFile)
			if err != nil {
				return fmt.Errorf("parse dotenv %q: %w", *dotenvFile, err)
			}
			store := dotenv.NewStore(vars)
			lookup = dotenv.ChainLookup(lookup, store.Lookup)
		}
		if err := envmacro.ExpandNode(node, envmacro.ExpandOptions{Lookup: lookup}); err != nil {
			return fmt.Errorf("env macro expansion: %w", err)
		}
	}

	// Marshal back to canonical YAML.
	yamlBytes, err := yamlparse.MarshalNode(node)
	if err != nil {
		return fmt.Errorf("marshal YAML: %w", err)
	}

	// Determine output path.
	outPath := inPath
	if *out != "" {
		outPath, err = resolveOutputPath(*out, "config.yml")
		if err != nil {
			return err
		}
	}

	if *mkdir {
		if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
			return fmt.Errorf("create directories for %q: %w", outPath, err)
		}
	}

	if err := os.WriteFile(outPath, yamlBytes, 0o644); err != nil {
		return fmt.Errorf("write %q: %w", outPath, err)
	}

	fmt.Fprintf(os.Stderr, "formatted %s\n", outPath)
	return nil
}
