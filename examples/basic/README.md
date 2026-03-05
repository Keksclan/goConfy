# Basic Example

This example demonstrates goConfy's core features: YAML config loading with environment macros, strict decoding into typed structs, and secret redaction.

## What It Shows

- Loading a YAML config file with `{ENV:KEY:default}` macros
- Strict decoding into a typed Go struct (including nested structs and `types.Duration`)
- Secret redaction: the `password` field (tagged `secret:"true"`) is replaced with `[REDACTED]` in output

## Files

- `main.go` — Go application that loads and prints config
- `config.yml` — YAML config with macro placeholders

## Running

From the repository root:

```bash
go run ./examples/basic
```

Expected output:

```text
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

## Customizing

Set environment variables to override defaults:

```bash
APP_PORT=3000 DB_DSN=postgres://prod:5432/mydb go run ./examples/basic
```
