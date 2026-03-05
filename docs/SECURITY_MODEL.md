# Security Model

This document explains the security design decisions in goConfy and provides recommended production policies.

## Exact-Match Macro Expansion

### Why `{ENV:KEY}` and `{FILE:PATH}` Instead of `${VAR}`

goConfy uses a custom macro format:
- Environment: `{ENV:KEY}` or `{ENV:KEY:default}`
- File: `{FILE:/path/to/file}` or `{FILE:/path/to/file:default}`

The macro is **only expanded when it is the entire YAML scalar value**. The regexes enforce this:

```regex
# ENV
^\{ENV:([A-Z0-9_]+)(?::([^}]*))?\}$

# FILE
^\{FILE:([^:]+)(?::([^}]*))?\}$
```

This is a deliberate security choice:

1. **No partial expansion**: A value like `http://{ENV:HOST}:8080` or `key={FILE:/tmp/key}` is **not** expanded. This prevents:
   - Accidental string interpolation bugs
   - Injection attacks where a macro resolves to a value containing another macro
   - Ambiguity about what is expanded and what isn't

2. **No recursive expansion**: The expanded value is used as-is, never re-scanned for macros.

3. **No nesting**: You cannot nest macros (e.g., `{ENV:{FILE:/path}:default}`).

4. **No shell-style variable references**: `$VAR`, `${VAR}`, and `$(cmd)` are never interpreted. This eliminates an entire class of shell injection vulnerabilities.

5. **Uppercase-only ENV keys**: The ENV regex only matches `[A-Z0-9_]+`, preventing accidental expansion of lowercase YAML values that happen to look like macros.

### What This Means in Practice

```yaml
# ✅ Expanded — entire value is a macro
port: "{ENV:PORT:8080}"
db_password: "{FILE:/run/secrets/db_password}"

# ❌ NOT expanded — macro is embedded in a string
url: "http://{ENV:HOST}:8080/path"
config: "source={FILE:/etc/config}"

# ❌ NOT expanded — shell-style variable
port: "${PORT}"

# ❌ NOT expanded — lowercase ENV key
port: "{ENV:port:8080}"
```

## File Macro Security

The `{FILE:PATH}` macro reads the entire content of the specified file, trims leading/trailing whitespace, and uses it as the configuration value.

- **Isolation**: File reading is performed by the process running goConfy. Ensure the process has the minimum necessary filesystem permissions.
- **Error Handling**: If a file is missing or unreadable, goConfy returns an error unless a default value is provided in the macro.
- **Redaction**: Values loaded via `{FILE:PATH}` are treated like any other configuration value and will be redacted if the target field is marked as a secret.

## Dotenv Does Not Mutate OS Environment

When loading a `.env` file, goConfy:

1. Parses the file into an in-memory key-value map
2. Uses the map as a lookup source for macro expansion
3. **Never calls `os.Setenv()`**

This means:
- The `.env` file cannot affect other parts of your application
- Other goroutines or libraries see only the real OS environment
- There are no race conditions from concurrent environment mutation
- The behavior is predictable and testable

### Precedence Control

By default, OS environment variables take precedence over `.env` values:

```go
goconfy.WithDotEnvOSPrecedence(true)  // default: OS wins
goconfy.WithDotEnvOSPrecedence(false) // .env wins
```

This allows production deployments to override `.env` defaults via real environment variables (e.g., from Kubernetes secrets).

## Environment Key Restrictions

### Prefix Filtering

```go
goconfy.WithEnvPrefix("MYAPP_")
```

When set, only environment variables starting with the prefix are eligible for macro expansion. This prevents accidental access to sensitive system variables like `PATH`, `HOME`, or `AWS_SECRET_ACCESS_KEY`.

### Allowlist

```go
goconfy.WithAllowedEnvKeys([]string{"DB_HOST", "DB_PORT", "DB_PASSWORD"})
```

When set, only the listed keys are expanded. Any macro referencing an unlisted key falls back to its default value (or empty string). This provides defense-in-depth for security-sensitive deployments.

### Combined

Prefix and allowlist can be used together. A key must satisfy **both** constraints to be expanded.

### Secret Redaction

### Struct Tag Redaction

Fields tagged with `secret:"true"` are automatically redacted:

```go
type Config struct {
    Password string `yaml:"password" secret:"true"`
}
```

When using `goconfy.Redacted()` or `goconfy.DumpRedactedJSON()`, the field value is replaced with `"[REDACTED]"`.

### Redaction by Convention (Opt-in)

You can enable automatic redaction for fields that match common naming conventions (e.g., `password`, `secret`, `token`, `key`, `private`).

```go
// Globally for the Load pipeline
goconfy.Load[Config](goconfy.WithRedactByConvention(true))

// Locally for a Redacted() call
goconfy.Redacted(cfg, goconfy.WithRedactByConventionOption(true))
```

Matching is case-insensitive and checks if the field name contains any of the keywords.

### Dot-Path Redaction

Additional paths can be redacted at runtime:

```go
goconfy.Redacted(cfg, goconfy.WithRedactPaths([]string{"db.password", "api.token"}))
```

### Generator Behavior

The `goconfygen init` command **never** outputs real secret values in generated templates:
- If a field has `secret:"true"` and an `env` tag, it outputs `{ENV:KEY:}` (empty default)
- If a field has `secret:"true"` and a `default` tag, the default is ignored
- A `# secret` comment is added above the field

### Redaction Guarantees

**What is guaranteed:**
- `Redacted()` and `DumpRedactedJSON()` replace tagged/pathed fields with `"[REDACTED]"`
- The generator never writes secret defaults to YAML or `.env` files
- Redaction works on nested structs, maps, and slices

**What is NOT guaranteed:**
- Redaction does not affect the original struct — it creates a copy
- If you `fmt.Printf("%+v", cfg)` directly, secrets are visible
- Redaction does not scrub log output from third-party libraries
- If a secret value is stored in a non-secret field, it won't be redacted

### Recommended Production Policies

1. **Tag all sensitive fields** with `secret:"true"` — passwords, tokens, API keys, certificates
2. **Always use `DumpRedactedJSON()`** for logging config at startup
3. **Never use `fmt.Printf`** or `log.Printf` with `%+v` on config structs
4. **Use `WithEnvPrefix()`** to limit which environment variables are accessible
5. **Use `WithAllowedEnvKeys()`** in high-security environments
6. **Keep `.env` files out of version control** (add to `.gitignore`)
7. **Use `WithDotEnvOSPrecedence(true)`** (the default) so Kubernetes/Docker secrets override `.env` defaults
8. **Validate configs in CI** with `goconfygen validate` to catch errors before deployment

## Strict Decoding

By default, goConfy uses strict YAML decoding (`KnownFields(true)`). This means:
- Typos in YAML keys cause errors instead of silent misconfiguration
- Removed config fields are caught immediately
- Extra keys from templates or copy-paste are rejected

To disable (not recommended for production):

```go
goconfy.WithStrictYAML(false)
```
