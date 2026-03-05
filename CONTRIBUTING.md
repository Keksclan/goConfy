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
# Build core packages
go build ./...

# Build optional tools module
(cd tools && go build ./...)

# Build core examples
go build ./examples/basic
go build ./examples/dotenv

# Build tool examples
(cd tools && go build ./examples/generator)
```

## Repository Hygiene

To keep the repository clean and secure, please follow these rules:

- **No Binaries**: Never commit compiled binaries, executables, or build artifacts (e.g., `*.exe`, `*.so`, `*.dylib`, `*.test`, `*.out`).
- **Source Only**: The repository should only contain source code, documentation, and configuration files.
- **Build Locally**: Generate builds and test binaries locally or let the CI handle them.
- **Clean Commits**: Ensure your `.gitignore` is up to date and that no temporary files or IDE-specific settings are committed.
- **Tools**: If you need to build tools in the `tools/` directory, use `go build` or `go install`. Do not commit the resulting binaries.

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
