# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased]

### Added
- **YAML Benchmarks**: Added `internal/bench/yaml_bench_test.go` to monitor YAML parsing and decoding performance.

### Changed
- **Minimum Go Version**: Lowered minimum Go version from 1.26 to 1.22.
- **YAML Parser Review**: Evaluated `github.com/goccy/go-yaml` for performance. Decided to stay with `gopkg.in/yaml.v3` to preserve strict `yaml.Node` AST compatibility (line/column info, macro expansion, and profile merging logic).

### Added
- **Full-Check & Release-Readiness**: Central `make verify` umbrella target for fmt, tests, race detector, and vulncheck.
- **Integration Test Suite**: New black-box integration tests in `tests/integration` covering the full pipeline (Load -> Merge -> Expand -> Normalize -> Validate -> Redact -> Explain).
- **MultiError Support**: Validation errors are now aggregated into a `MultiError` for better feedback.
- **Enhanced Error Reporting**: `FieldError` now includes `Layer`, `Path`, `Line`, and `Column` information where applicable.
- **Deterministic Generator**: Golden tests in `tools/tests` ensure deterministic output for `goconfygen`.
- **Release Checklist**: Added `docs/RELEASE_CHECKLIST.md` for standardized releases.

### Changed
- **Repository Hardening**: Removed committed binaries from `tools/`, updated `.gitignore` with standard Go rules, and added a CI guard to prevent future binary commits.
- **Tool Splitting**: Tools moved to a separate Go module (`tools/`) to keep core dependencies minimal.
- **Dotenv Security**: Loading `.env` files no longer mutates the global `os` environment, respecting the Security Model.
- **Macro Restrictions**: Macros are now only expanded when the scalar is an exact match (no partial interpolation), preventing unintended side effects.

### Fixed
- **Macro Redaction**: Secrets expanded via macros are now correctly redacted in `DumpRedactedJSON` and the Explain report.

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
