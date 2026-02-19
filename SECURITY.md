# Security Policy

## Supported Versions

Only the latest version of goConfy is supported for security updates.

## Reporting a Vulnerability

Please report security vulnerabilities to issues ig i dont have a standart do report these in the moment sry.

## Security Model Summary

- **No automatic env scanning**: only explicitly defined `{ENV:KEY:default}` macros are expanded
- **Exact-match only**: no `${VAR}`, no inline expansion, no recursive resolution
- **Strict decoding**: rejects unknown YAML fields by default
- **Secret redaction**: `secret:"true"` tag + dot-path redaction for safe logging
- **Dotenv isolation**: `.env` values never injected into `os.Environ()`
- **Dotenv precedence**: OS env wins by default (`WithDotEnvOSPrecedence(true)`)
- **Env key restrictions**: prefix filtering and explicit allowlists supported

For the full security model with detailed explanations and recommended production policies, see [docs/SECURITY_MODEL.md](docs/SECURITY_MODEL.md).
