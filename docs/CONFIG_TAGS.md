# Config Struct Tags

goConfy uses Go struct tags to control YAML mapping, environment variable binding, documentation, and security behavior. The `goconfygen` CLI reads these tags to generate config templates with appropriate values and comments.

## Tag Reference

### `yaml`

**Used by:** loader, generator

Controls the YAML key name for a field. Standard `gopkg.in/yaml.v3` behavior.

```go
type Config struct {
    Host string `yaml:"host"`
    Port int    `yaml:"port"`
}
```

Special values:
- `yaml:"-"` — field is skipped entirely (not in YAML, not generated)
- `yaml:",inline"` — inline the struct fields into the parent

### `env`

**Used by:** generator (for macro output), loader (via macro expansion)

Specifies the environment variable key used in macro expansion.

```go
type Config struct {
    Port int `yaml:"port" env:"APP_PORT"`
}
```

**Generator behavior:** When `env` is set, the generated YAML value becomes a macro:

```yaml
# env: APP_PORT
port: "{ENV:APP_PORT:8080}"
```

The default portion of the macro comes from the `default` tag (if present).

### `default`

**Used by:** generator

Sets the default value for a field. Used as the fallback in `{ENV:KEY:DEFAULT}` macros, or as the literal value when no `env` tag is present.

```go
type Config struct {
    Port     int    `yaml:"port"     default:"8080"`
    LogLevel string `yaml:"log_level" default:"info"`
}
```

**Generator output without `env`:**

```yaml
port: 8080
log_level: info
```

**Generator output with `env`:**

```yaml
port: "{ENV:APP_PORT:8080}"
log_level: "{ENV:LOG_LEVEL:info}"
```

If no `default` tag is set, the generator uses type-appropriate zero values:
- `string` → `""`
- `int` → `0`
- `bool` → `false`
- `time.Duration` / `types.Duration` → `0s`

### `desc`

**Used by:** generator

Provides a human-readable description emitted as a YAML comment.

```go
type Config struct {
    Port int `yaml:"port" desc:"HTTP server listen port"`
}
```

**Generator output:**

```yaml
# desc: HTTP server listen port
port: 8080
```

### `example`

**Used by:** generator

Shows an example value as a YAML comment. Useful for complex or non-obvious formats.

```go
type Config struct {
    DSN string `yaml:"dsn" example:"postgres://user:pass@host:5432/db"`
}
```

**Generator output:**

```yaml
# example: postgres://user:pass@host:5432/db
dsn: ""
```

### `required`

**Used by:** generator

Marks a field as required. Emitted as a comment hint.

```go
type Config struct {
    DSN string `yaml:"dsn" required:"true"`
}
```

**Generator output:**

```yaml
# required: true
dsn: ""
```

> **Note:** The `required` tag is currently a documentation hint only. goConfy does not enforce it at load time — use the `Validate()` hook for runtime validation.

### `secret`

**Used by:** loader (redaction), generator

Marks a field as containing sensitive data.

```go
type Config struct {
    Password string `yaml:"password" secret:"true"`
}
```

**Loader behavior:**
- `Redacted()` and `DumpRedactedJSON()` replace the value with `"[REDACTED]"`

**Generator behavior:**
- Default values are **never** written (even if `default` tag exists)
- Output uses empty-default macro: `{ENV:DB_PASSWORD:}`
- A `# secret` comment is added

```yaml
# secret
# env: DB_PASSWORD
password: "{ENV:DB_PASSWORD:}"
```

### `sep`

**Used by:** generator (comment only)

Specifies the separator for parsing comma-separated environment values into slices.

```go
type Config struct {
    AllowedOrigins []string `yaml:"allowed_origins" env:"ALLOWED_ORIGINS" sep:","`
}
```

**Generator output:**

```yaml
# env: ALLOWED_ORIGINS
# sep: ,
allowed_origins: "{ENV:ALLOWED_ORIGINS:}"
```

> **Note:** The `sep` tag is currently a documentation hint for the generator. Actual slice parsing from env values is not yet implemented in the loader.

## Complete Example

```go
type Config struct {
    // Server settings
    Host string `yaml:"host" default:"0.0.0.0"   desc:"Bind address"      env:"APP_HOST"`
    Port int    `yaml:"port" default:"8080"       desc:"HTTP listen port"  env:"APP_PORT" example:"3000"`

    // Database
    DB struct {
        DSN      string `yaml:"dsn"      env:"DB_DSN"      desc:"Connection string" required:"true" example:"postgres://localhost:5432/mydb"`
        Password string `yaml:"password" env:"DB_PASSWORD" desc:"Database password" secret:"true"`
        MaxConns int    `yaml:"max_conns" default:"10"      desc:"Connection pool size"`
    } `yaml:"db"`

    // Feature flags
    Debug   bool     `yaml:"debug"   default:"false" desc:"Enable debug mode" env:"DEBUG"`
    Origins []string `yaml:"origins" env:"ALLOWED_ORIGINS" sep:"," desc:"CORS allowed origins"`
}
```

**Generated YAML:**

```yaml
# desc: Bind address
# env: APP_HOST
host: "{ENV:APP_HOST:0.0.0.0}"
# desc: HTTP listen port
# env: APP_PORT
# example: 3000
port: "{ENV:APP_PORT:8080}"
db:
  # desc: Connection string
  # required: true
  # env: DB_DSN
  # example: postgres://localhost:5432/mydb
  dsn: "{ENV:DB_DSN:}"
  # secret
  # desc: Database password
  # env: DB_PASSWORD
  password: "{ENV:DB_PASSWORD:}"
  # desc: Connection pool size
  max_conns: 10
# desc: Enable debug mode
# env: DEBUG
debug: "{ENV:DEBUG:false}"
# desc: CORS allowed origins
# env: ALLOWED_ORIGINS
# sep: ,
origins: "{ENV:ALLOWED_ORIGINS:}"
```

## Tag Interaction Matrix

| Tag | Loader | Generator | Redacted() |
|-----|--------|-----------|------------|
| `yaml` | ✅ Key mapping | ✅ Key name | — |
| `env` | — (via macros) | ✅ Macro output | — |
| `default` | — | ✅ Default value | — |
| `desc` | — | ✅ Comment | — |
| `example` | — | ✅ Comment | — |
| `required` | — | ✅ Comment | — |
| `secret` | — | ✅ Redacts default | ✅ Redacts value |
| `sep` | — | ✅ Comment | — |
