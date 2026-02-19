# goConfy

goConfy is a strongly typed, strict, and secure YAML configuration loader for Go.

It provides a complete configuration pipeline: YAML parsing → environment macro expansion → dotenv loading → profile-based overrides → strict decoding → normalization → validation → secret redaction.

## Features

- **YAML parsing** via `gopkg.in/yaml.v3`
- **Environment macros**: `{ENV:KEY:default}` — exact-match, secure, no shell injection
- **Dotenv support**: load `.env` files without mutating `os.Environ()`
- **Profile-based overrides**: `dev`, `staging`, `prod` merged into base config
- **Strict decoding**: rejects unknown YAML keys (catches typos)
- **Typed structs**: decode into `int`, `bool`, `string`, `time.Duration`, nested structs
- **Normalization hook**: `Normalize()` called after decode
- **Validation hook**: `Validate()` called after normalization
- **Secret redaction**: `secret:"true"` tag + dot-path redaction for safe logging
- **Generator CLI** (`goconfygen`): generate YAML templates, validate, format, and dump configs
- **Interactive TUI** (`goconfytui`): browse configs, preview, validate, format, and dump interactively

## Installation

```bash
go get github.com/keksclan/goConfy@latest
```

## Quickstart

### 1. Define a Config Struct

```go
package main

import (
    "fmt"
    "log"

    goconfy "github.com/keksclan/goConfy"
    "github.com/keksclan/goConfy/types"
)

type Config struct {
    Host    string         `yaml:"host"`
    Port    int            `yaml:"port"`
    Timeout types.Duration `yaml:"timeout"`
    DB      struct {
        DSN      string `yaml:"dsn"`
        Password string `yaml:"password" secret:"true"`
    } `yaml:"db"`
}

func main() {
    cfg, err := goconfy.Load[Config](
        goconfy.WithFile("config.yml"),
    )
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    fmt.Printf("Host: %s, Port: %d\n", cfg.Host, cfg.Port)

    // Safe logging — secrets are replaced with [REDACTED]
    json, _ := goconfy.DumpRedactedJSON(cfg)
    fmt.Println(json)
}
```

### 2. Create config.yml

```yaml
host: localhost
port: "{ENV:APP_PORT:8080}"
timeout: 30s
db:
  dsn: "{ENV:DB_DSN:postgres://localhost:5432/mydb}"
  password: "{ENV:DB_PASSWORD:}"
```

### 3. Create .env (Optional)

```dotenv
APP_PORT=3000
DB_DSN=postgres://db.prod:5432/mydb
DB_PASSWORD=supersecret
```

### 4. Run

```bash
go run main.go
```

Output:

```
Host: localhost, Port: 8080
{
  "db": {
    "dsn": "postgres://localhost:5432/mydb",
    "password": "[REDACTED]"
  },
  "host": "localhost",
  "port": 8080,
  "timeout": "30s"
}
```

## Macros

goConfy uses exact-match environment macros:

```
{ENV:KEY}          → look up KEY, empty string if missing
{ENV:KEY:default}  → look up KEY, use "default" if missing
```

### Rules

- The macro must be the **entire** YAML value — no inline macros
- Key names must be uppercase + digits + underscores: `[A-Z0-9_]+`
- No recursive expansion — the resolved value is never re-scanned
- No shell-style `${VAR}` or `$(cmd)` — by design (see [Security Model](docs/SECURITY_MODEL.md))

### Examples

```yaml
# ✅ Correct — entire value is a macro
port: "{ENV:PORT:8080}"
host: "{ENV:HOST:localhost}"

# ✅ No default — empty string if HOST is not set
host: "{ENV:HOST}"

# ❌ Will NOT expand — macro is embedded in a string
url: "http://{ENV:HOST}:8080"

# ❌ Will NOT expand — shell-style
port: "${PORT}"
```

### Pitfalls

- **No inline macros**: `"prefix-{ENV:KEY}-suffix"` is NOT expanded. Use separate fields or compose in your application code.
- **No quoting inside defaults**: `{ENV:KEY:"value"}` — the quotes become part of the default. Use `{ENV:KEY:value}` instead.

## Dotenv

Load a `.env` file as an additional variable source for macro expansion. The file is parsed into memory and **never injected into `os.Environ()`**.

### Basic Usage

```go
cfg, err := goconfy.Load[Config](
    goconfy.WithFile("config.yml"),
    goconfy.WithDotEnvFile(".env"),          // enables dotenv + sets path
)
```

### Precedence

By default, **OS environment wins** over `.env` values:

```go
goconfy.WithDotEnvOSPrecedence(true)   // default — OS env wins
goconfy.WithDotEnvOSPrecedence(false)  // .env wins over OS env
```

### Optional Missing File

By default, a missing `.env` file causes an error:

```go
goconfy.WithDotEnvOptional(true)  // silently ignore missing .env
```

### Supported .env Format

```dotenv
# Comments are ignored
KEY=value
export ANOTHER_KEY=value
QUOTED="double quoted with \n escapes"
LITERAL='single quoted, no escapes'
```

| Syntax | Result |
|--------|--------|
| `KEY=value` | `value` (trimmed) |
| `KEY=""` | empty string |
| `KEY='  '` | two spaces (preserved) |
| `KEY="a#b"` | `a#b` (hash is literal in quotes) |
| `KEY=a #comment` | `a` (inline comment stripped) |

## Profiles

Profiles allow environment-specific overrides within a single YAML file.

### Usage

```yaml
host: localhost
port: 8080
profiles:
  prod:
    host: 0.0.0.0
    port: 443
```

```go
cfg, err := goconfy.Load[Config](
    goconfy.WithFile("config.yml"),
    goconfy.WithEnableProfiles(true),
)
```

Set the active profile:

```bash
export APP_PROFILE=prod
```

Or explicitly:

```go
goconfy.WithProfile("prod")
```

### Merge Rules

- Scalars: profile value replaces base
- Mappings: merged recursively (only overridden keys change)
- Sequences: profile replaces entire list
- The `profiles` key is removed before decoding

See [docs/PROFILES.md](docs/PROFILES.md) for full details.

## Redaction

### Secret Tag

```go
type Config struct {
    Password string `yaml:"password" secret:"true"`
}

cfg := Config{Password: "secret123"}
redacted := goconfy.Redacted(cfg)
// password → "[REDACTED]"
```

### Dot-Path Redaction

```go
redacted := goconfy.Redacted(cfg, goconfy.WithRedactPaths([]string{"db.password"}))
```

### DumpRedactedJSON

```go
json, err := goconfy.DumpRedactedJSON(cfg)
// Returns JSON with all secrets replaced by "[REDACTED]"
```

## Hooks

### Normalize()

If your config struct implements `Normalize()`, it is called after decoding:

```go
func (c *Config) Normalize() {
    if c.Host == "" {
        c.Host = "localhost"
    }
    c.LogLevel = strings.ToLower(c.LogLevel)
}
```

### Validate()

If your config struct implements `Validate() error`, it is called after normalization:

```go
func (c *Config) Validate() error {
    if c.Port < 1 || c.Port > 65535 {
        return fmt.Errorf("invalid port: %d", c.Port)
    }
    return nil
}
```

## Generator CLI (goconfygen)

`goconfygen` generates YAML config templates, validates configs, formats YAML, and dumps redacted output — all driven by typed Go config structs.

### Install

```bash
go install github.com/keksclan/goConfy/cmd/goconfygen@latest
# or build locally:
go build -o goconfygen ./cmd/goconfygen
```

### Registry Approach

Because Go cannot import arbitrary packages at runtime, `goconfygen` uses a registry pattern. Your project registers its config type:

```go
package config

import "github.com/keksclan/goConfy/gen/registry"

type configProvider struct{}

func (configProvider) ID() string { return "myservice" }
func (configProvider) New() any   { return &Config{} }

func init() {
    registry.Register(configProvider{})
}
```

### Commands

#### `goconfygen init` — Generate YAML template

```bash
goconfygen init -id myservice -out config.yml
goconfygen init -id myservice -out config.yml -profile dev,prod -dotenv
```

#### `goconfygen validate` — Validate config

```bash
goconfygen validate -id myservice -in config.yml
goconfygen validate -id myservice -in config.yml -dotenv .env -profile prod -print
```

#### `goconfygen fmt` — Format YAML

```bash
goconfygen fmt -in config.yml
goconfygen fmt -in config.yml -expand -dotenv .env -out resolved.yml
```

#### `goconfygen dump` — Dump redacted JSON

```bash
goconfygen dump -id myservice -in config.yml -dotenv .env
```

### Struct Tags for Generator

| Tag | Purpose | Example |
|-----|---------|---------|
| `yaml:"key"` | YAML key name | `yaml:"port"` |
| `env:"KEY"` | Environment macro | `env:"APP_PORT"` |
| `default:"val"` | Default value | `default:"8080"` |
| `desc:"text"` | Description comment | `desc:"HTTP port"` |
| `example:"val"` | Example comment | `example:"3000"` |
| `required:"true"` | Required hint | `required:"true"` |
| `secret:"true"` | Secret (redacted) | `secret:"true"` |
| `sep:","` | Slice separator | `sep:","` |

See [docs/CONFIG_TAGS.md](docs/CONFIG_TAGS.md) for full reference.
See [docs/CLI.md](docs/CLI.md) for complete CLI reference.

## Examples

### examples/basic

Demonstrates YAML loading with macros and strict decoding:

```bash
go run ./examples/basic
```

### examples/dotenv

Demonstrates `.env` file integration with typed decoding:

```bash
go run ./examples/dotenv
```

### examples/generator

Demonstrates the registry provider and goconfygen usage:

```bash
# See examples/generator/README.md for details
go build ./cmd/goconfygen
```

### examples/tui

Sample config and .env for exploring the TUI:

```bash
go run ./cmd/goconfytui
# Then open examples/tui/config.yml and examples/tui/.env
# See examples/tui/README.md for a full walkthrough
```

## goconfytui (TUI)

`goconfytui` is an interactive terminal UI for goConfy, providing the same workflows as the CLI in a keyboard-driven interface powered by [Charm](https://charm.sh/) (bubbletea + lipgloss + bubbles).

### Install / Build

```bash
# Install globally
go install github.com/keksclan/goConfy/cmd/goconfytui@latest

# Or build locally
go build -o goconfytui ./cmd/goconfytui
```

### Features

- **Inspect config** — browse RAW YAML, EXPANDED, MERGED, and REDACTED JSON tabs
- **Validate** — run the full pipeline and see VALID / INVALID with error details
- **Format** — preview before/after and write with confirmation
- **Dump** — view redacted JSON output
- **Init** — generate config templates from registry providers
- **Settings** — toggle strict mode, dotenv options, profile env var, redaction paths
- **Profile handling** — see active profile, available profiles, and override inline
- **Directory browser** — navigate and select files with Ctrl+B

### Key Bindings

| Key | Action |
|-----|--------|
| `q` / `esc` | Back / quit (from home) |
| `↑` / `↓` | Navigate |
| `enter` | Select / confirm |
| `tab` | Switch tabs / fields |
| `ctrl+s` | Save / write |
| `r` | Reload from disk |
| `v` | Validate |
| `f` | Format |
| `d` | Dump |
| `i` | Init / template |
| `?` | Toggle help |
| `ctrl+b` | Open directory browser |
| `ctrl+c` | Force quit |

### Security

- Secrets are **never** shown in plaintext.
- YAML previews use best-effort dot-path redaction (configurable in Settings).
- The REDACTED JSON tab always uses `secret:"true"` struct tags.
- Default redacted paths: `redis.password`, `auth.opaque.client_secret`, `postgres.url`.

See [examples/tui/README.md](examples/tui/README.md) for a step-by-step walkthrough.

## Documentation

| Document | Description |
|----------|-------------|
| [Getting Started](docs/GETTING_STARTED.md) | Full tutorial from zero to config |
| [CLI Reference](docs/CLI.md) | All goconfygen commands, flags, examples |
| [Config Tags](docs/CONFIG_TAGS.md) | All supported struct tags |
| [Profiles](docs/PROFILES.md) | Profile behavior and merge semantics |
| [Security Model](docs/SECURITY_MODEL.md) | Security design decisions |

## Compatibility

- **Go 1.26+** required
- Core dependency: `gopkg.in/yaml.v3`
- TUI dependencies: `github.com/charmbracelet/bubbletea`, `lipgloss`, `bubbles`

## License

See [LICENSE](LICENSE) file.
