# Generator & TUI (optional tools module)

`goConfy` keeps runtime loading (`Load`, merging, validation, redaction) in the core module.
Generator and TUI tooling is optional and lives in the separate module under `tools/`.

## Install

```bash
# Generator CLI
go install github.com/keksclan/goConfy/tools/cmd/goconfygen@latest

# Interactive TUI
go install github.com/keksclan/goConfy/tools/cmd/goconfytui@latest
```

## Build locally from this repository

```bash
mkdir -p tools/bin
(cd tools && go build -o ../tools/bin/goconfygen ./cmd/goconfygen)
(cd tools && go build -o ../tools/bin/goconfytui ./cmd/goconfytui)
```

## Registry import path

For provider registration in your project, use:

```go
import "github.com/keksclan/goConfy/tools/generator/registry"
```

## Commands

- `goconfygen init` — generate YAML template (+ optional profiles and `.env`)
- `goconfygen validate` — validate typed config from YAML
- `goconfygen fmt` — canonical YAML formatting (+ optional macro expansion)
- `goconfygen dump` — print redacted JSON output

Full command reference: [CLI.md](CLI.md)

## TUI

Run locally:

```bash
(cd tools && go run ./cmd/goconfytui)
```

Example walkthrough and sample files:

- [tools/examples/tui/README.md](../tools/examples/tui/README.md)
