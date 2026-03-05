# goConfy

goConfy is a strongly typed, strict, and secure YAML configuration loader for Go.

It provides a complete configuration pipeline: YAML parsing → environment macro expansion → dotenv loading → profile-based overrides → strict decoding → normalization → validation → secret redaction.

## Features

- **YAML parsing** via `gopkg.in/yaml.v3`
- **Macros**: `{ENV:KEY:default}` and `{FILE:/path:default}` — exact-match, secure, no shell injection
- **Dotenv support**: load `.env` files without mutating `os.Environ()`
- **Profile-based overrides**: `dev`, `staging`, `prod` merged into base config
- **Strict decoding**: rejects unknown YAML keys (catches typos)
- **Typed structs**: decode into `int`, `bool`, `string`, `time.Duration`, nested structs
- **Normalization hook**: `Normalize()` called after decode
- **Validation hook**: `Validate()` called after normalization
- **Secret redaction**: `secret:"true"` tag, **redaction-by-convention**, and dot-path redaction
- **Optional tooling module**: `goconfygen` (CLI) and `goconfytui` (TUI) are available separately in `./tools`

## Installation

```bash
go get github.com/keksclan/goConfy@latest
```

This installs the **core runtime loader** only.

Optional generator/TUI tooling lives in a separate module at `./tools`.
See [docs/GENERATOR.md](docs/GENERATOR.md) and [docs/INSTALL_TOOLS.md](docs/INSTALL_TOOLS.md).

## Versioning & Compatibility

Wir folgen [SemVer](https://semver.org/).

- **v0.x.x**: Experimentelle Phase. Breaking Changes an der API sind jederzeit möglich.
- **v1.x.x**: Stabile API. Wir garantieren Rückwärtskompatibilität innerhalb einer Major-Version.

Weitere Informationen zum Release-Prozess finden Sie in [RELEASE.md](RELEASE.md).

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

## Optional Tools (separate module)

`goconfygen` (CLI) and `goconfytui` (TUI) are intentionally split from the core runtime.

Install directly:

```bash
go install github.com/keksclan/goConfy/tools/cmd/goconfygen@latest
go install github.com/keksclan/goConfy/tools/cmd/goconfytui@latest
```

Or build locally from this repository:

```bash
(cd tools && go build -o goconfygen ./cmd/goconfygen)
(cd tools && go build -o goconfytui ./cmd/goconfytui)
```

Provider registry import path for generator tooling:

```go
import "github.com/keksclan/goConfy/tools/generator/registry"
```

See [docs/GENERATOR.md](docs/GENERATOR.md) for tool usage and [docs/CLI.md](docs/CLI.md) for full CLI flags.

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

### tools/examples/generator

Demonstrates the registry provider and goconfygen usage:

```bash
# See tools/examples/generator/README.md for details
(cd tools && go build ./cmd/goconfygen)
```

### tools/examples/tui

Sample config and .env for exploring the TUI:

```bash
(cd tools && go run ./cmd/goconfytui)
# Then open tools/examples/tui/config.yml and tools/examples/tui/.env
# See tools/examples/tui/README.md for a full walkthrough
```

## Documentation

| Document | Description |
|----------|-------------|
| [Getting Started](docs/GETTING_STARTED.md) | Full tutorial from zero to config |
| [Generator & TUI](docs/GENERATOR.md) | Optional tooling module (`goconfygen`, `goconfytui`) |
| [CLI Reference](docs/CLI.md) | All goconfygen commands, flags, examples |
| [Install Tools](docs/INSTALL_TOOLS.md) | Core-only vs tools installation and build commands |
| [Config Tags](docs/CONFIG_TAGS.md) | All supported struct tags |
| [Profiles](docs/PROFILES.md) | Profile behavior and merge semantics |
| [Security Model](docs/SECURITY_MODEL.md) | Security design decisions |

## Compatibility

- **Go 1.26+** required
- Core dependency: `gopkg.in/yaml.v3`
- Tool dependencies (only in `tools` module): `bubbletea`, `lipgloss`, `bubbles`

## License

See [LICENSE](LICENSE) file.
