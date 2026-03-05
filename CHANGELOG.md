# Changelog

All notable changes to this project will be documented in this file.

## [0.2.0] - 2026-02-19

### Added
- **Explain Reporting**: Optional trace reporting for configuration values to track their origin (base, profile, env, etc.) without leaking secrets.
  - `WithExplainReporter(func(explain.Report))`: Option to enable reporting.
  - `explain.Report`: Structured report with path, source, overrides, and redacted values.
  - Support for text and JSON report output.
- **goconfygen CLI**: generator tool for YAML config templates, validation, formatting, and redacted dump
  - `goconfygen init`: generate YAML templates from registered Go config types
  - `goconfygen validate`: validate YAML against typed config structs
  - `goconfygen fmt`: canonicalize YAML formatting with optional macro expansion
  - `goconfygen dump`: dump resolved config as redacted JSON
- **Registry package** (`gen/registry`): type registry for CLI type discovery via `Provider` interface
- **YAML template generator** (`gen/yamltemplate`): reflection-based template emission with macro strings, comments, and profile skeletons
- **Extended struct tags**: `default`, `desc`, `example`, `required`, `env`, `sep` for generator documentation
- **Comprehensive documentation**:
  - `docs/GETTING_STARTED.md`: full tutorial from clone to production
  - `docs/CLI.md`: complete CLI reference with all commands, flags, and workflows
  - `docs/SECURITY_MODEL.md`: detailed security design decisions and production policies
  - `docs/CONFIG_TAGS.md`: all supported struct tags with examples
  - `docs/PROFILES.md`: profile behavior and merge semantics
- **GoDoc**: package-level `doc.go` for all packages, GoDoc on all exported symbols
- **Examples**: `examples/basic`, `examples/dotenv`, `examples/generator` with READMEs

## [0.1.0] - 2026-02-19
- Initial release of goConfy.
