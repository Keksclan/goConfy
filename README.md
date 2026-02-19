# goConfy

goConfy is a strongly typed, strict, and secure YAML configuration loader for Go.

## Features

- YAML parsing via `yaml.v3`
- Exact environment macro expansion: `{ENV:KEY:default}`
- Profile-based overrides
- Strict decoding into typed structs
- Validation and normalization hooks
- Secret redaction helpers
- Fully modular internal architecture

## Installation

```bash
go get github.com/keksclan/goConfy@latest
```

## Quickstart

```go
package main

import (
	"fmt"
	"github.com/keksclan/goConfy"
)

type Config struct {
	Port int `yaml:"port"`
	Host string `yaml:"host"`
}

func main() {
	cfg, err := goconfy.Load[Config](
		goconfy.WithFile("config.yml"),
	)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Config: %+v\n", cfg)
}
```

## Macro Syntax

Macros are expanded only if the string exactly matches the pattern:
`^\{ENV:([A-Z0-9_]+)(?::([^}]*))?\}$`

Example:
```yaml
port: "{ENV:PORT:8080}"
```
This will look for the `PORT` environment variable and use `8080` if it's not set.

## Profiles

Profiles allow you to override base configuration for different environments (e.g., `dev`, `prod`).
Enable with `WithEnableProfiles(true)`.
The default profile is read from the `APP_PROFILE` environment variable.

```yaml
app:
  port: 8080
profiles:
  prod:
    app:
      port: 80
```

## Security

- No automatic environment scanning.
- Only expand macros explicitly defined in YAML.
- Support for `secret:"true"` struct tag to redact sensitive data.
- Dot-path redaction supported.

## Redaction Example

```go
type Config struct {
	Password string `yaml:"password" secret:"true"`
}

cfg := Config{Password: "secret123"}
redacted := goconfy.Redacted(cfg)
// Output: {Password: [REDACTED]}
```

## Dotenv Support

goConfy can optionally load a `.env` file and use it as an additional environment variable source for macro expansion. The `.env` file is **never** injected into the OS environment.

### Basic Usage

```go
cfg, err := goconfy.Load[Config](
    goconfy.WithFile("config.yml"),
    goconfy.WithDotEnvFile(".env"),
)
```

Setting `WithDotEnvFile` automatically enables dotenv loading. You can also enable it explicitly:

```go
goconfy.WithDotEnvEnabled(true)
```

### Precedence

By default, OS environment variables take precedence over `.env` values. To reverse this:

```go
goconfy.WithDotEnvOSPrecedence(false) // .env wins over OS env
```

### Optional Missing File

By default, a missing `.env` file causes an error. To allow it:

```go
goconfy.WithDotEnvOptional(true)
```

### Supported .env Format

```dotenv
# Comments are ignored
KEY=value
export ANOTHER_KEY=value
QUOTED="double quoted with \n escapes"
LITERAL='single quoted, no escapes'
```

## Dotenv + Typed Decoding Example

A full working example lives in [`examples/dotenv/`](examples/dotenv/). It shows:

- YAML config with `{ENV:KEY:default}` macros
- A `.env` file providing overrides
- Typed decoding into `int`, `bool`, `string`, and `types.Duration` fields

Run it from the repo root:

```bash
go run ./examples/dotenv
```

### Quoting Rules

The `.env` parser handles quotes predictably:

| Syntax | Result |
|--------|--------|
| `KEY=value` | `value` (trimmed) |
| `KEY=""` | empty string |
| `KEY=''` | empty string |
| `KEY="  "` | two spaces (preserved) |
| `KEY='  '` | two spaces (preserved) |
| `KEY="a#b"` | `a#b` (hash is literal in quotes) |
| `KEY=a #comment` | `a` (inline comment stripped) |
| `KEY=a#b` | `a#b` (no space before `#`, not a comment) |

Single quotes are literal (no escapes). Double quotes support `\n`, `\t`, `\r`, `\"`, `\\`.

## Duration Type

`goconfy` provides a custom `Duration` type that supports human-readable YAML parsing.

```go
type Config struct {
	Timeout goconfy.Duration `yaml:"timeout"`
}
```
YAML: `timeout: 5m`

## goconfygen CLI

`goconfygen` is a CLI tool that generates YAML config templates, validates configs, formats YAML, and dumps redacted config output — all driven by your typed Go config structs.

### Installation

```bash
go install github.com/keksclan/goConfy/cmd/goconfygen@latest
```

### Registry Approach

Because Go cannot import arbitrary packages at runtime, `goconfygen` uses a **registry** pattern. Your project registers its config type, then the CLI operates on it by ID.

#### 1. Define your config struct with tags

```go
type Config struct {
    Host    string `yaml:"host" default:"0.0.0.0" desc:"Bind address" env:"APP_HOST" example:"localhost"`
    Port    int    `yaml:"port" default:"8080" desc:"HTTP port" env:"APP_PORT"`
    Debug   bool   `yaml:"debug" default:"false" desc:"Enable debug mode"`
    DB      DBConfig `yaml:"db"`
}

type DBConfig struct {
    DSN      string `yaml:"dsn" default:"postgres://localhost/mydb" desc:"Connection string" env:"DB_DSN" required:"true"`
    Password string `yaml:"password" secret:"true" env:"DB_PASSWORD" desc:"Database password"`
}
```

#### 2. Register a provider

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

#### 3. Use the CLI

```bash
# Generate YAML config template with defaults and comments
goconfygen init -id myservice -out ./config.yml

# Include profile skeleton and .env file
goconfygen init -id myservice -out ./config.yml -profile dev -dotenv -force

# Validate a config file
goconfygen validate -id myservice -in ./config.yml

# Format/canonicalize YAML
goconfygen fmt -in ./config.yml

# Dump redacted config as JSON
goconfygen dump -id myservice -in ./config.yml
```

### Commands

| Command    | Purpose                                              |
|------------|------------------------------------------------------|
| `init`     | Generate YAML template from registered config type   |
| `validate` | Validate YAML by decoding into typed struct          |
| `fmt`      | Canonicalize YAML formatting                         |
| `dump`     | Show resulting config as redacted JSON               |

### Struct Tags Reference

| Tag        | Example               | Purpose                                   |
|------------|-----------------------|-------------------------------------------|
| `yaml`     | `yaml:"host"`        | YAML key name                             |
| `default`  | `default:"8080"`     | Default value for template generation     |
| `desc`     | `desc:"HTTP port"`   | Description emitted as YAML comment       |
| `env`      | `env:"APP_PORT"`     | Environment variable source               |
| `secret`   | `secret:"true"`      | Mark as secret (never output in clear)    |
| `required` | `required:"true"`    | Mark as required (comment + hint)         |
| `example`  | `example:"localhost"`| Example value shown in comment            |
| `sep`      | `sep:","`            | Separator for slice env parsing           |

### Generated Output

Fields with `env` tags produce macro values: `"{ENV:APP_PORT:8080}"`.
Secret fields never include real defaults: `"{ENV:DB_PASSWORD:}"`.
Comments include `desc`, `env`, `required`, `secret`, and `example` info.

See the full example in [`examples/generator/`](examples/generator/).

## License

MIT License
