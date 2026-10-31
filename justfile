# Default recipe executed when running `just` without arguments
set positional-arguments := true

default: build

# List available recipes
help:
    @just --list

# ------------------------------------------------------------------------------
# Variables & Configuration
# ------------------------------------------------------------------------------
BIN_DIR := "bin"
DIST_DIR := "dist"
RELEASE_TARGETS := "linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64"

# ------------------------------------------------------------------------------
# Build Targets
# ------------------------------------------------------------------------------

# Build the unified public CLI and WASM assets
build: build-cli build-wasm

# Build everything (public CLI, WASM, and maintainer tool)
build-local: build build-internal

# Build the public-facing CLI binary
build-cli:
	@echo "Building public CLI binary..."
	@mkdir -p {{BIN_DIR}}
	go build -o {{BIN_DIR}}/prosetta ./cmd/prosetta
	@echo "Public CLI binary ready in {{BIN_DIR}}/prosetta"

# Build the standalone maintainer binary
build-internal:
	@echo "Building maintainer CLI binary..."
	@mkdir -p {{BIN_DIR}}
	go build -o {{BIN_DIR}}/generate ./cmd/generate
	@echo "Maintainer CLI binary ready in {{BIN_DIR}}/generate"

# Package the metadata index, common templates, and each root category
build-bundle:
	go run ./cmd/generate bundle --output={{DIST_DIR}}/problems

# Build the Go WebAssembly core and its matching JavaScript runtime bridge
build-wasm:
	@echo "Building WebAssembly engine target..."
	@mkdir -p {{DIST_DIR}}
	GOOS=js GOARCH=wasm go build -o {{DIST_DIR}}/prosetta.wasm ./cmd/wasm
	@cp "$(go env GOROOT)/lib/wasm/wasm_exec.js" {{DIST_DIR}}/wasm_exec.js
	@echo "WASM target and runtime bridge ready in {{DIST_DIR}}/"

# ------------------------------------------------------------------------------
# Workspace Generation Shortcuts
# ------------------------------------------------------------------------------

# Scaffold a new category (e.g. just category --id=sorting --title=Sorting)
category *args:
    go run ./cmd/generate category "$@"

# Scaffold a new concept (e.g. just concept --parent=algorithms --id=two-sum --title="Two Sum")
concept *args:
    go run ./cmd/generate concept "$@"

# Find local concept paths and run IDs
list *args:
    go run ./cmd/generate list "$@"

# ------------------------------------------------------------------------------
# Execution Shortcuts
# ------------------------------------------------------------------------------

# Probe local language toolchains
probe *args:
    go run ./cmd/prosetta probe "$@"

# Run public cases (e.g. just run --id=c-12345678 --lang=go --stdin < solution.go)
run *args:
    go run ./cmd/prosetta run "$@"

# Download and activate published catalog content
catalog *args:
    go run ./cmd/prosetta catalog "$@"

# Run full evaluation against all cases (public + hidden)
submit *args:
    go run ./cmd/prosetta run --all "$@"

# ------------------------------------------------------------------------------
# Development & Testing
# ------------------------------------------------------------------------------

# Run all Go unit tests
test:
    @echo "Running unit tests..."
    go test -v ./...

# Run Go WASM unit and integration tests through Go's Node runtime wrapper
test-wasm:
	GOOS=js GOARCH=wasm go test -v -timeout=90s -exec="env -i PATH=$(dirname "$(command -v node)"):/usr/bin:/bin $(go env GOROOT)/lib/wasm/go_js_wasm_exec" ./cmd/wasm ./internal/wasm

# Run Go tests with the race detector
test-race:
    @echo "Running unit tests with the race detector..."
    go test -race ./...

# Run race tests with CI coverage settings and generate local reports
test-cover:
    @echo "Running unit tests with the race detector and coverage..."
    go test -race -covermode=atomic -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated at coverage.html"

# Run go vet and static analysis
lint:
    @echo "Running static analysis..."
    go vet ./...

# Local vet, race tests, and coverage reports; CI enforces the coverage minimum
check-go: lint test-cover

# Versioned CLI binaries for Linux, macOS, and Windows on amd64 and arm64
release-cli:
    #!/usr/bin/env sh
    set -eu
    version="$(go run ./cmd/prosetta version)"
    mkdir -p "{{BIN_DIR}}"
    for target in {{RELEASE_TARGETS}}; do
        target_os="${target%/*}"
        target_arch="${target#*/}"
        extension=""
        if [ "$target_os" = windows ]; then
            extension=".exe"
        fi
        output="{{BIN_DIR}}/prosetta-${version}-${target_os}-${target_arch}${extension}"
        echo "Building $output..."
        CGO_ENABLED=0 GOOS="$target_os" GOARCH="$target_arch" go build -trimpath -o "$output" ./cmd/prosetta
    done

# All platform CLI binaries, WASM core, and runtime bridge for v* tags
release: release-cli build-wasm

# Build catalog assets for branch-based GitHub Pages publishing
release-catalog: build-bundle

# Tidy and verify go.mod / go.sum dependencies
tidy:
    @echo "Tidying Go module dependencies..."
    go mod tidy
    go mod verify

# Format all Go source files
format:
    @echo "Formatting code..."
    go fmt ./...

# ------------------------------------------------------------------------------
# Cleanup
# ------------------------------------------------------------------------------

# Remove all build output and test coverage files
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf {{BIN_DIR}}/ {{DIST_DIR}}/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"
