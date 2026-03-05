# Getting Started with goConfy

This tutorial walks you through using goConfy from scratch: loading configs, using macros, dotenv files, profiles, redaction, and the generator CLI.

## Prerequisites

- **Go 1.26+** installed
- A terminal (bash, PowerShell, or similar)

## 1. Clone the Repository

```bash
git clone https://github.com/keksclan/goConfy.git
cd goConfy
```

## 2. Run Tests

Verify everything works:

```bash
go test ./...
```

Expected output:

```
ok      github.com/keksclan/goConfy/tests    1.5s
```

All other packages will show `[no test files]` — that's expected.

## 3. Run the Basic Example

The `examples/basic` example demonstrates YAML loading with macros and strict decoding.

```bash
go run ./examples/basic
```

Expected output (values depend on your environment):

```
Host: localhost
Port: 8080
Timeout: 30s
Redacted: map[db:map[dsn:postgres://localhost:5432/mydb password:[REDACTED]] host:localhost port:8080 timeout:30s]
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

Notice how `password` is replaced with `[REDACTED]` because it has the `secret:"true"` tag.

## 4. Run the Dotenv Example

The `examples/dotenv` example shows how `.env` files integrate with YAML macros.

```bash
go run ./examples/dotenv
```

This loads `examples/dotenv/config.yml` and resolves macros using environment variables and the `.env` file.

## 5. Create Your Own Config

### Step 1: Define a Config Struct

Create a new Go file (`main.go`):

```go
package main

import (
    "fmt"
    "log"

    goconfy "github.com/keksclan/goConfy"
)

type Config struct {
    Host     string `yaml:"host"`
    Port     int    `yaml:"port"     env:"APP_PORT"    default:"8080"`
    LogLevel string `yaml:"log_level" env:"LOG_LEVEL"   default:"info"`
    DB       struct {
        DSN      string `yaml:"dsn"      env:"DB_DSN"`
        Password string `yaml:"password" env:"DB_PASSWORD" secret:"true"`
    } `yaml:"db"`
}

func main() {
    cfg, err := goconfy.Load[Config](
        goconfy.WithFile("config.yml"),
        goconfy.WithDotEnvFile(".env"),
        goconfy.WithDotEnvOptional(true),
    )
    if err != nil {
        log.Fatalf("failed to load config: %v", err)
    }

    fmt.Printf("Host: %s\n", cfg.Host)
    fmt.Printf("Port: %d\n", cfg.Port)

    // Safe logging with redaction
    json, _ := goconfy.DumpRedactedJSON(cfg)
    fmt.Println(json)
}
```

### Step 2: Create config.yml

```yaml
host: localhost
port: "{ENV:APP_PORT:8080}"
log_level: "{ENV:LOG_LEVEL:info}"
db:
  dsn: "{ENV:DB_DSN:postgres://localhost:5432/mydb}"
  password: "{ENV:DB_PASSWORD:}"
```

### Step 3: Create .env (Optional)

```dotenv
APP_PORT=3000
LOG_LEVEL=debug
DB_DSN=postgres://db.example.com:5432/prod
DB_PASSWORD=supersecret
```

### Step 4: Run

```bash
go run main.go
```

Output:

```
Host: localhost
Port: 3000
{
  "db": {
    "dsn": "postgres://db.example.com:5432/prod",
    "password": "[REDACTED]"
  },
  "host": "localhost",
  "log_level": "debug",
  "port": 3000
}
```

## 6. Generate Config with goconfygen

The `goconfygen` CLI can generate YAML templates from your Go struct types.

### Build the CLI

```bash
(cd tools && go build -o ../goconfygen ./cmd/goconfygen)
```

### Register Your Config Type

In your project, create a provider:

```go
package config

import "github.com/keksclan/goConfy/tools/generator/registry"

type Config struct {
    Host string `yaml:"host" default:"localhost" desc:"Server hostname"`
    Port int    `yaml:"port" env:"APP_PORT"      default:"8080" desc:"HTTP port"`
    DB   struct {
        DSN      string `yaml:"dsn"      env:"DB_DSN"      desc:"Database connection string"`
        Password string `yaml:"password" env:"DB_PASSWORD"  secret:"true" desc:"Database password"`
    } `yaml:"db"`
}

type configProvider struct{}

func (configProvider) ID() string { return "myservice" }
func (configProvider) New() any   { return &Config{} }

func init() {
    registry.Register(configProvider{})
}
```

### Generate a YAML Template

```bash
./goconfygen init -id myservice -out config.yml
```

This generates `config.yml` with defaults, macros, and comments.

### Validate a Config

```bash
./goconfygen validate -id myservice -in config.yml
```

Exit code 0 means the config is valid.

### Dump Redacted Config

```bash
./goconfygen dump -id myservice -in config.yml
```

Prints the resolved config as redacted JSON.

## 7. Using Profiles

Add a `profiles` section to your YAML:

```yaml
host: localhost
port: 8080
profiles:
  prod:
    host: prod.example.com
    port: 443
  staging:
    host: staging.example.com
    port: 8443
```

Enable profiles in your loader:

```go
cfg, err := goconfy.Load[Config](
    goconfy.WithFile("config.yml"),
    goconfy.WithEnableProfiles(true),
)
```

Set the active profile via environment variable:

```bash
export APP_PROFILE=prod
go run main.go
# host will be "prod.example.com", port will be 443
```

Or set it explicitly:

```go
goconfy.WithProfile("prod")
```

## 8. Using the TUI

goConfy provides an interactive TUI (`goconfytui`) in the optional `tools` module that provides the same
operations as the CLI in a keyboard-driven terminal interface.

### Install

```bash
go install github.com/keksclan/goConfy/tools/cmd/goconfytui@latest
# or build locally
(cd tools && go build -o ../goconfytui ./cmd/goconfytui)
```

### Quick Start

```bash
./goconfytui
```

1. Select **Open / Inspect config** from the menu.
2. Type the path to your YAML file (e.g. `config.yml`).
3. Optionally enter a `.env` file path.
4. Press `enter` — the Inspect screen shows four tabs:
   - **RAW YAML** — file contents (secrets redacted)
   - **EXPANDED** — after macro expansion (secrets redacted)
   - **MERGED** — after profile merge (secrets redacted)
   - **REDACTED JSON** — typed config with `[REDACTED]` for secrets
5. Press `tab` to cycle tabs, `esc` to go back.
6. From the home menu use `3` to validate, `4` to format, `5` to dump, `6` for settings.
7. Press `?` at any time for a help overlay.

See [tools/examples/tui/README.md](../tools/examples/tui/README.md) for a full walkthrough with sample files.

## Troubleshooting

### "strict YAML decode: unknown field"

You have a key in your YAML that doesn't match any struct field. Either:
- Fix the YAML key name
- Add the field to your struct
- Disable strict mode: `goconfy.WithStrictYAML(false)`

### Profile not applied

- Ensure `WithEnableProfiles(true)` is set
- Check that `APP_PROFILE` env var matches a key under `profiles:` in your YAML
- Profile names are case-sensitive

### Dotenv file not found

By default, a missing `.env` file is an error. Use:
```go
goconfy.WithDotEnvOptional(true)
```

### Macro not expanding

Macros must be the **entire** YAML value, using exact-match format:
```yaml
# ✅ Correct — entire value is a macro
port: "{ENV:PORT:8080}"

# ❌ Wrong — macro embedded in a string (will NOT expand)
url: "http://{ENV:HOST}:8080"
```

### Duration / Port / Bool type errors

Ensure YAML values match Go types:
- Duration: `"30s"`, `"5m"`, `"1h30m"` (Go duration format)
- Int: `8080` (no quotes needed in YAML, but macros produce strings that are auto-converted)
- Bool: `true` / `false`

## Next Steps

- See [CLI Reference](CLI.md) for all goconfygen commands and flags
- See [Config Tags](CONFIG_TAGS.md) for all supported struct tags
- See [Profiles](PROFILES.md) for profile merge semantics
- See [Security Model](SECURITY_MODEL.md) for security design decisions
