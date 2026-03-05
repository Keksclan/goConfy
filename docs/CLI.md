# goconfygen CLI Reference

`goconfygen` is a CLI tool for generating, validating, formatting, and dumping YAML configuration files driven by typed Go config structs.

## Installation

```bash
# Install from source
go install github.com/keksclan/goConfy/tools/cmd/goconfygen@latest

# Or build locally
mkdir -p tools/bin
(cd tools && go build -o ../tools/bin/goconfygen ./cmd/goconfygen)
```

## Overview

```
goconfygen <command> [flags]

Commands:
  init       Generate a YAML config template from a registered config type
  validate   Validate an existing YAML config file
  fmt        Canonicalize YAML formatting
  dump       Dump resolved config as redacted JSON
```

Running `goconfygen` without arguments prints usage information.

> **TUI alternative:** All of the above operations are also available interactively
> via `goconfytui`. See [docs/GENERATOR.md](GENERATOR.md) or
> run `(cd tools && go run ./cmd/goconfytui)`.

---

## Registry-Based Type Discovery

Because Go cannot import arbitrary packages at runtime, `goconfygen` uses a **registry** pattern. Your project registers its config type, and the CLI operates on it by ID.

### How It Works

1. Your project implements `registry.Provider`:

```go
package config

import "github.com/keksclan/goConfy/tools/generator/registry"

type configProvider struct{}

func (configProvider) ID() string { return "myservice" }
func (configProvider) New() any   { return &Config{} }

func init() {
    registry.Register(configProvider{})
}
```

2. The CLI looks up the provider by `-id` flag:

```bash
goconfygen init -id myservice -out config.yml
```

3. The registry uses the `New()` method to create a zero-valued instance of your config type, then reflects on it to generate templates, validate, or dump.

### Provider Interface

```go
type Provider interface {
    ID() string   // Unique identifier for this config type
    New() any     // Returns a pointer to a new zero-valued config struct
}
```

---

## Commands

### `goconfygen init`

Generate a YAML config template from a registered config type.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-id` | string | (required) | Registry ID of the config type |
| `-pkg` | string | | Go package import path (reserved; not yet supported) |
| `-type` | string | | Go type name (reserved; not yet supported) |
| `-out` | string | `./config.yml` | Output file path or directory |
| `-profile` | string | | Comma-separated profile names for skeleton generation |
| `-dotenv` | bool | `false` | Also generate a sample `.env` file |
| `-force` | bool | `false` | Overwrite existing files |
| `-mkdir` | bool | `true` | Create parent directories if needed |

**Examples:**

```bash
# Basic generation
goconfygen init -id myservice

# Custom output path
goconfygen init -id myservice -out ./deploy/config.yml

# With profile skeleton and .env file
goconfygen init -id myservice -profile dev,staging,prod -dotenv

# Overwrite existing files
goconfygen init -id myservice -out config.yml -force
```

**Output format:**

The generated YAML includes:
- Keys derived from `yaml` struct tags
- Values as `{ENV:KEY:DEFAULT}` macros when `env` tag is present
- Default values from `default` tag (or type-appropriate zero values)
- Comments from `desc`, `required`, `env`, `example` tags
- Secret fields output as `{ENV:KEY:}` with `# secret` comment (never real values)

**Exit codes:**
- `0`: Success
- `1`: Error (missing flags, registry lookup failure, file exists without `-force`)

---

### `goconfygen validate`

Validate an existing YAML config file by decoding into the registered type.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-id` | string | (required) | Registry ID of the config type |
| `-in` | string | (required) | Input YAML file path |
| `-dotenv` | string | | Path to `.env` file for macro expansion |
| `-profile` | string | | Profile name to apply |
| `-strict` | bool | `true` | Reject unknown YAML fields |
| `-print` | bool | `false` | Print redacted JSON on success |

**Examples:**

```bash
# Basic validation
goconfygen validate -id myservice -in config.yml

# With dotenv and profile
goconfygen validate -id myservice -in config.yml -dotenv .env -profile prod

# Print redacted output on success
goconfygen validate -id myservice -in config.yml -print

# Relaxed mode (allow unknown fields)
goconfygen validate -id myservice -in config.yml -strict=false
```

**Behavior:**
1. Reads the YAML file
2. Expands environment macros (using OS env + optional dotenv)
3. Applies profile override (if `-profile` is set)
4. Decodes into the typed struct (strict or relaxed)
5. Calls `Normalize()` if implemented
6. Calls `Validate()` if implemented
7. Optionally prints redacted JSON (if `-print`)

**Exit codes:**
- `0`: Config is valid
- `1`: Validation error (message printed to stderr)

---

### `goconfygen fmt`

Canonicalize YAML formatting.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-in` | string | (required) | Input YAML file path |
| `-out` | string | (same as `-in`) | Output file path |
| `-expand` | bool | `false` | Expand macros before formatting |
| `-dotenv` | string | | Path to `.env` file (used with `-expand`) |

**Examples:**

```bash
# Format in place
goconfygen fmt -in config.yml

# Format to a new file
goconfygen fmt -in config.yml -out config.formatted.yml

# Expand macros and format
goconfygen fmt -in config.yml -expand -dotenv .env -out config.resolved.yml
```

**Behavior:**
- Parses YAML into a node tree
- Optionally expands environment macros
- Re-marshals with consistent formatting (2-space indent, sorted keys)
- Writes output (overwrites input by default)

**Exit codes:**
- `0`: Success
- `1`: Error (file not found, parse error)

---

### `goconfygen dump`

Show the resolved, typed config as redacted JSON for debugging.

**Flags:**

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `-id` | string | (required) | Registry ID of the config type |
| `-in` | string | (required) | Input YAML file path |
| `-dotenv` | string | | Path to `.env` file |
| `-profile` | string | | Profile name to apply |

**Examples:**

```bash
# Basic dump
goconfygen dump -id myservice -in config.yml

# With dotenv
goconfygen dump -id myservice -in config.yml -dotenv .env

# With profile
goconfygen dump -id myservice -in config.yml -profile prod
```

**Output:**

Redacted JSON to stdout. Fields with `secret:"true"` are shown as `"[REDACTED]"`.

```json
{
  "host": "localhost",
  "port": 8080,
  "db": {
    "dsn": "postgres://localhost:5432/mydb",
    "password": "[REDACTED]"
  }
}
```

**Exit codes:**
- `0`: Success
- `1`: Error

---

## Common Workflows

### New Service Bootstrap

```bash
# 1. Register your config type (in your Go code)
# 2. Generate initial config
goconfygen init -id myservice -out config.yml -dotenv -profile dev,staging,prod

# 3. Edit config.yml with your actual values/macros
# 4. Validate
goconfygen validate -id myservice -in config.yml
```

### CI Validation

Add to your CI pipeline:

```yaml
# GitHub Actions example
- name: Validate config
  run: |
    mkdir -p tools/bin
    (cd tools && go build -o ../tools/bin/goconfygen ./cmd/goconfygen)
    tools/bin/goconfygen validate -id myservice -in config.yml -dotenv .env.example
```

```bash
# Exit code 0 = valid, 1 = invalid
goconfygen validate -id myservice -in config.yml -strict
echo $?
```

### Formatting Configs in a Repo

```bash
# Format all config files consistently
goconfygen fmt -in config.yml
goconfygen fmt -in config.staging.yml
goconfygen fmt -in config.prod.yml

# Check in CI that formatting is canonical
goconfygen fmt -in config.yml -out /tmp/formatted.yml
diff config.yml /tmp/formatted.yml
```

### Debugging Config Resolution

```bash
# See what the final config looks like after macro expansion
DB_PASSWORD=test123 goconfygen dump -id myservice -in config.yml -dotenv .env -profile prod
```
