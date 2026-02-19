# Dotenv + Typed Decoding Example

This example demonstrates how goConfy loads a YAML config with `{ENV:KEY:default}` macros,
resolves values from a `.env` file, and decodes them into a fully typed Go struct.

## Files

- `config.yml` — YAML config using environment macros
- `.env` — dotenv file providing overrides
- `main.go` — loads config and prints a redacted JSON dump

## Run

From the repository root:

```bash
go run ./examples/dotenv
```

## What It Shows

- **String** fields (`app.name`) are resolved from `.env`
- **Boolean** fields (`app.version_display`) are correctly decoded from string macros
- **Integer** fields (`server.grpc.port`) are correctly decoded from string macros
- **Duration** fields (`server.grpc.timeout`) are parsed via the custom `types.Duration` type
- Secrets tagged with `secret:"true"` would be redacted in the JSON dump
