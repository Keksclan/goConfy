# TUI Example

This directory contains sample files for exploring the goConfy TUI.

## Files

| File         | Description                                    |
|-------------|------------------------------------------------|
| `config.yml` | Sample YAML config with env macros and profiles |
| `.env`       | Sample dotenv file with secret values           |

## Running the TUI

From the project root:

```bash
(cd tools && go run ./cmd/goconfytui)
```

## Walkthrough

1. **Start the TUI** — you will see the Workflow Menu.
2. **Press `1`** (or arrow down + enter) to select **Open / Inspect config**.
3. **Enter paths** in the file picker:
   - Config YAML: `tools/examples/tui/config.yml`
   - .env file: `tools/examples/tui/.env`
   - Output path: leave empty
4. **Press `enter`** to open the Inspect screen.
5. **Press `tab`** to cycle through preview tabs:
   - **RAW YAML** — original file with secrets redacted
   - **EXPANDED** — env macros replaced with values (secrets redacted)
   - **MERGED** — after profile merge (secrets redacted)
   - **REDACTED JSON** — typed config with `[REDACTED]` for secrets
6. **Press `esc`** to go back to the home menu.
7. **Press `3`** to validate — enter the same paths and press `v`.
8. **Press `4`** to format — toggle options and press `f` to preview.
9. **Press `5`** to dump redacted JSON.
10. **Press `6`** to change settings (strict mode, profile env var, etc.).
11. **Press `?`** at any time to toggle the global help bar.
12. **Press `q`** from the home screen to quit.

## Notes

- Secrets are **never** displayed in plaintext. YAML previews use best-effort
  dot-path redaction; the REDACTED JSON tab is always safe.
- The TUI works without mouse — all navigation uses keyboard shortcuts.
- Profile info is shown on the inspect screen (available profiles + active one).
