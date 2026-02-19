# Profiles

Profiles allow you to define environment-specific configuration overrides (e.g., `dev`, `staging`, `prod`) within a single YAML file.

## Overview

A profile is a named set of overrides stored under a top-level `profiles` key in your YAML config. When a profile is selected, its values are deep-merged into the base configuration before decoding.

## Enabling Profiles

```go
cfg, err := goconfy.Load[Config](
    goconfy.WithFile("config.yml"),
    goconfy.WithEnableProfiles(true),
)
```

Profiles are disabled by default. You must explicitly enable them.

## Selecting a Profile

### Via Environment Variable (Default)

The default environment variable is `APP_PROFILE`:

```bash
export APP_PROFILE=prod
go run main.go
```

### Custom Environment Variable

```go
goconfy.WithProfileEnvVar("MY_PROFILE")
```

### Explicit Selection

```go
goconfy.WithProfile("prod")
```

Explicit selection takes precedence over the environment variable.

## YAML Structure

```yaml
# Base configuration (always applied)
host: localhost
port: 8080
db:
  dsn: postgres://localhost:5432/mydb
  max_conns: 5
  password: "{ENV:DB_PASSWORD:}"

# Profile overrides
profiles:
  dev:
    port: 3000
    db:
      dsn: postgres://localhost:5432/devdb
      max_conns: 2

  staging:
    host: staging.example.com
    port: 8443
    db:
      dsn: postgres://staging-db:5432/mydb
      max_conns: 10

  prod:
    host: 0.0.0.0
    port: 443
    db:
      dsn: "{ENV:DB_DSN:}"
      max_conns: 50
```

## Merge Semantics

Profile values are deep-merged into the base configuration:

1. **Scalar values** (string, int, bool): profile value **replaces** base value
2. **Mapping values** (nested structs): merged recursively — only specified keys are overridden
3. **Sequence values** (arrays/slices): profile value **replaces** the entire sequence (no element-level merge)
4. **Missing keys**: base values are preserved if the profile doesn't override them

### Example

Base config:
```yaml
host: localhost
port: 8080
db:
  dsn: postgres://localhost:5432/mydb
  max_conns: 5
```

Profile `prod`:
```yaml
profiles:
  prod:
    port: 443
    db:
      max_conns: 50
```

Result after merge (with `APP_PROFILE=prod`):
```yaml
host: localhost       # unchanged — not in profile
port: 443             # overridden by profile
db:
  dsn: postgres://localhost:5432/mydb  # unchanged — not in profile
  max_conns: 50       # overridden by profile
```

## Profile Removal

After the profile is applied, the `profiles` key is **removed** from the YAML tree before decoding. This means:
- Your config struct does not need a `Profiles` field
- Strict mode won't reject the `profiles` key
- The decoded struct only contains the final merged values

## No Profile Selected

If no profile is selected (no env var set, no explicit profile), the base config is used as-is. The `profiles` key is still removed before decoding.

## Unknown Profile

If the selected profile name does not exist under `profiles`, a non-fatal behavior occurs:
- The base config is used unchanged
- No error is raised (the profile is simply not found)

## Generator Support

The `goconfygen init` command can generate a profiles skeleton:

```bash
goconfygen init -id myservice -profile dev,staging,prod
```

This adds a `profiles` section with empty override blocks:

```yaml
host: "{ENV:APP_HOST:localhost}"
port: "{ENV:APP_PORT:8080}"
profiles:
  dev: {}
  staging: {}
  prod: {}
```

You can then fill in the overrides as needed.

## Best Practices

1. **Keep base config as the development default** — developers should be able to run locally without setting `APP_PROFILE`
2. **Use macros in profiles** — profile values can contain `{ENV:KEY:DEFAULT}` macros too
3. **Don't duplicate everything** — only override what changes per environment
4. **Use CI validation** — validate each profile in CI:
   ```bash
   goconfygen validate -id myservice -in config.yml -profile dev
   goconfygen validate -id myservice -in config.yml -profile prod
   ```
