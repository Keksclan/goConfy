.PHONY: test test-race lint fmt-check vulncheck help

# Default target
all: test lint

## test: Run all tests
test:
	go test ./...

## test-race: Run all tests with race detector
test-race:
	go test -race ./...

## lint: Run golangci-lint
lint:
	golangci-lint run ./...

## fmt-check: Check formatting without modifying files
fmt-check:
	@if [ -n "$$(gofmt -l .)" ]; then \
		echo "Following files are not formatted:"; \
		gofmt -l .; \
		exit 1; \
	fi

## vulncheck: Run govulncheck
vulncheck:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed -e 's/## //g' -e 's/: /	/g'
