# Generator Example

This example shows how to use `goconfygen` with the registry approach.

## Setup

1. Define your config struct with tags (`yaml`, `default`, `desc`, `env`, `secret`, `required`, `example`).
2. Create a `Provider` and register it via `registry.Register()` in an `init()` function.

See [register.go](register.go) for the full example.

## Generated Files

- [config.yml](config.yml) — YAML template with defaults, env macros, and comments
- [.env](.env) — Sample `.env` file with all env-sourced fields

## Running

```bash
# Generate config template programmatically
go run ./examples/generator/

# Or using the CLI (after building goconfygen with your provider linked in):
# goconfygen init -id myservice -out ./config.yml -profile dev -dotenv -force
```

## Tags Reference

| Tag          | Example                     | Purpose                                 |
|--------------|-----------------------------|-----------------------------------------|
| `yaml`       | `yaml:"host"`              | YAML key name                           |
| `default`    | `default:"8080"`           | Default value                           |
| `desc`       | `desc:"HTTP port"`         | Description (YAML comment)              |
| `env`        | `env:"APP_PORT"`           | Environment variable source             |
| `secret`     | `secret:"true"`            | Mark as secret (redacted in output)     |
| `required`   | `required:"true"`          | Mark as required                        |
| `example`    | `example:"localhost"`      | Example value (shown in comment)        |
| `sep`        | `sep:","`                  | Separator for slice env parsing         |
