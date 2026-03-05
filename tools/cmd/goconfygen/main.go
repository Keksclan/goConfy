// Command goconfygen is a CLI tool for generating YAML config templates,
// validating configs, formatting YAML, and dumping redacted config output.
//
// Usage:
//
//	goconfygen <command> [flags]
//
// Commands:
//
//	init       Generate a YAML config template from a registered config type
//	validate   Validate a YAML config file
//	fmt        Canonicalize YAML formatting
//	dump       Dump redacted config as JSON
package main

import (
	"fmt"
	"os"

	"github.com/keksclan/goConfy/tools/internal/tools/generator/commands"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cmd := os.Args[1]
	args := os.Args[2:]

	var err error
	switch cmd {
	case "init":
		err = commands.RunInit(args)
	case "validate":
		err = commands.RunValidate(args)
	case "fmt":
		err = commands.RunFmt(args)
	case "dump":
		err = commands.RunDump(args)
	case "help", "-h", "--help":
		printUsage()
		return
	default:
		fmt.Fprintf(os.Stderr, "goconfygen: unknown command %q\n\n", cmd)
		printUsage()
		os.Exit(1)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "goconfygen %s: %v\n", cmd, err)
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprint(os.Stderr, `Usage: goconfygen <command> [flags]

Commands:
  init       Generate a YAML config template from a registered config type
  validate   Validate a YAML config file
  fmt        Canonicalize YAML formatting
  dump       Dump redacted config as JSON

Run 'goconfygen <command> -h' for command-specific help.
`)
}
