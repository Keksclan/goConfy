.PHONY: test test-race fmt-check vulncheck verify help

# Default target
all: verify

## verify: Run all checks (fmt, test, race, vulncheck, build)
verify: fmt-check test test-race vulncheck build-tools

## build-tools: Build tool binaries
build-tools:
	mkdir -p tools/bin
	cd tools && go build -o bin/goconfygen ./cmd/goconfygen
	cd tools && go build -o bin/goconfytui ./cmd/goconfytui

## test: Run all tests
test:
	go test ./...
	cd tools && go test ./...

## test-race: Run all tests with race detector
test-race:
	go test -race ./...
	cd tools && go test -race ./...


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
	cd tools && go run golang.org/x/vuln/cmd/govulncheck@latest ./...

## update-golden: Update golden files for tests
update-golden:
	go test ./tests -run TestGolden -update

## help: Show this help message
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^##' Makefile | sed -e 's/## //g' -e 's/: /	/g'
