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

## License

MIT License
