# Core-only vs Tools Installation

## Core-only usage (no generator dependencies)

Install only the runtime loader:

```bash
go get github.com/keksclan/goConfy@latest
```

This keeps your application on the minimal core dependency set.

## Optional tools installation

Install `goconfygen`:

```bash
go install github.com/keksclan/goConfy/tools/cmd/goconfygen@latest
```

Install `goconfytui`:

```bash
go install github.com/keksclan/goConfy/tools/cmd/goconfytui@latest
```

## Build tools from local checkout

From the repository root:

```bash
mkdir -p tools/bin
(cd tools && go build -o bin/goconfygen ./cmd/goconfygen)
(cd tools && go build -o bin/goconfytui ./cmd/goconfytui)
```

## Run tools without installing

```bash
(cd tools && go run ./cmd/goconfygen --help)
(cd tools && go run ./cmd/goconfytui)
```
