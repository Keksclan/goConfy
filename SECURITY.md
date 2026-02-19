# Security Policy

## Supported Versions

Only the latest version of goConfy is supported for security updates.

## Reporting a Vulnerability

Please report security vulnerabilities to keksiclan@gmail.com.

## Security Model

- **No automatic env scanning**: goConfy does not automatically load all environment variables into your configuration.
- **Explicit macros**: Only environment variables explicitly defined in YAML using the `{ENV:KEY:default}` syntax are expanded.
- **Strict decoding**: YAML decoding is strict by default, rejecting unknown fields to prevent accidental configuration errors.
- **Redaction**: Support for `secret:"true"` tags and manual redaction paths to prevent sensitive data from being logged or printed.
- **Dotenv isolation**: The `.env` file loader does **not** call `os.Setenv`. Values are kept in an internal store and used only as a lookup source for macro expansion. This prevents accidental mutation of the global OS environment.
- **Dotenv precedence**: By default, OS environment variables take precedence over `.env` values (`WithDotEnvOSPrecedence(true)`). This is the recommended production policy, ensuring that deployment-level environment variables always override file-based defaults.
- **Recommended production policy**: Use `WithDotEnvOptional(false)` (default) in production to fail fast if the expected `.env` file is missing. Use OS precedence to ensure runtime environment always wins.
