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

# ------------------------------------------------------------------------------
# Build Targets
# ------------------------------------------------------------------------------

# Build the unified public CLI and WASM assets
build: build-cli build-wasm

# Build everything (public CLI, WASM, and maintainer tools)
build-local: build build-internal

# Build the public-facing CLI binary
build-cli:
	@echo "Building public CLI binary..."
	@mkdir -p {{BIN_DIR}}
	go build -o {{BIN_DIR}}/prosetta ./cmd/prosetta
	@echo "Public CLI binary ready in {{BIN_DIR}}/prosetta"

# Build standalone binaries
build-internal:
	@echo "Building standalone CLI binaries..."
	@mkdir -p {{BIN_DIR}}
	go build -o {{BIN_DIR}}/probe ./cmd/probe
	go build -o {{BIN_DIR}}/run ./cmd/run
	go build -o {{BIN_DIR}}/generate ./cmd/generate
	@echo "CLI binaries ready in {{BIN_DIR}}/"

# Build WebAssembly binary target
build-wasm:
	@echo "Building WebAssembly engine target..."
	@mkdir -p {{DIST_DIR}}
	GOOS=js GOARCH=wasm go build -o {{DIST_DIR}}/prosetta.wasm ./cmd/wasm
	@GOROOT="$$(go env GOROOT)"; \
	if [ -n "$$GOROOT" ] && [ -f "$$GOROOT/lib/wasm/wasm_exec.js" ]; then \
		echo "Copying wasm_exec.js from GOROOT/lib/wasm..."; \
		cp "$$GOROOT/lib/wasm/wasm_exec.js" {{DIST_DIR}}/wasm_exec.js; \
	elif [ -n "$$GOROOT" ] && [ -f "$$GOROOT/misc/wasm/wasm_exec.js" ]; then \
		echo "Copying wasm_exec.js from GOROOT/misc/wasm..."; \
		cp "$$GOROOT/misc/wasm/wasm_exec.js" {{DIST_DIR}}/wasm_exec.js; \
	else \
		echo "Downloading official wasm_exec.js for Go..." && \
		curl -sSL https://raw.githubusercontent.com/golang/go/master/lib/wasm/wasm_exec.js -o {{DIST_DIR}}/wasm_exec.js; \
	fi
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

# Compile dist/index.json from problems/
index *args:
    go run ./cmd/generate index "$@"

# ------------------------------------------------------------------------------
# Execution Shortcuts
# ------------------------------------------------------------------------------

# Probe local language toolchains
probe *args:
    go run ./cmd/prosetta probe "$@"

# Run evaluation on public cases (e.g. just run --concept=hello-world --path=problems/.../go/main.go)
run *args:
    go run ./cmd/prosetta run "$@"

# Run full evaluation against all cases (public + hidden)
submit *args:
    go run ./cmd/prosetta run --all "$@"

# ------------------------------------------------------------------------------
# Development & Testing
# ------------------------------------------------------------------------------

# Run all unit tests
test:
    @echo "Running unit tests..."
    go test -v ./...

# Run unit tests with code coverage analysis
test-cover:
    @echo "Running unit tests with coverage..."
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated at coverage.html"

# Run go vet and static analysis
lint:
    @echo "Running static analysis..."
    go vet ./...

# Tidy and verify go.mod / go.sum dependencies
tidy:
    @echo "Tidying Go module dependencies..."
    go mod tidy
    go mod verify

# Format all Go source files
format:
    @echo "Formatting code..."
    go fmt ./...

# Configure local Git repository to use workspace hooks in .githooks/
setup-hooks:
    @echo "⚓ Setting up Git pre-commit hooks..."
    git config core.hooksPath .githooks
    chmod +x .githooks/pre-commit
    @echo "✓ Git hooks active!"

# ------------------------------------------------------------------------------
# Cleanup
# ------------------------------------------------------------------------------

# Remove built directories and test coverage files
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf {{BIN_DIR}}/ {{DIST_DIR}}/
	@rm -f coverage.out coverage.html
	@echo "Clean complete!"
