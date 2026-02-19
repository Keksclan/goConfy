# Contributing to goConfy

Thank you for your interest in contributing to goConfy!

## Prerequisites

- Go 1.26+
- Git

## Getting Started

```bash
git clone https://github.com/keksclan/goConfy.git
cd goConfy
go test ./...
```

## How to Contribute

1. Fork the repository.
2. Create a new branch for your changes.
3. Make your changes and add tests.
4. Run the checks below.
5. Submit a pull request.

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./tests/...

# Run a specific test
go test -v -run TestLoaderBasic ./tests/...
```

## Running Lint / Vet

```bash
# Run go vet
go vet ./...

# Format all code
gofmt -w .
```

## Building

```bash
# Build all packages
go build ./...

# Build the CLI
go build -o goconfygen ./cmd/goconfygen

# Build examples
go build ./examples/basic
go build ./examples/dotenv
```

## Code Style

- Follow standard Go idioms and conventions.
- Use `gofmt` for formatting.
- Every exported type, function, constant, and variable must have GoDoc comments.
- Each package must have a `doc.go` with package-level documentation.
- Non-trivial internal functions should have comments explaining behavior and edge cases.

## Testing Guidelines

- Add tests for new features and bug fixes.
- Tests live in the `tests/` directory.
- Use table-driven tests where appropriate.
- Ensure `go test ./...` passes before submitting.

## Documentation

- Update `README.md` if your change affects user-facing behavior.
- Update relevant docs in `docs/` for feature changes.
- Update `CHANGELOG.md` with a description of your change.
