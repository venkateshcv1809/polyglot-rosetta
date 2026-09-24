# Default recipe executed when running `just` without arguments
default: build

# List available recipes
help:
    @just --list

# ------------------------------------------------------------------------------
# Build Targets
# ------------------------------------------------------------------------------

# Build the unified CLI binary into ./bin directory
build: build-prosetta build-wasm

# Build the main unified CLI umbrella binary
build-prosetta:
	@echo "Building prosetta..."
	@mkdir -p bin
	go build -o bin/prosetta main.go

# Build WebAssembly binary target
build-wasm:
	@echo "Building WebAssembly engine target..."
	@mkdir -p dist
	GOOS=js GOARCH=wasm go build -o dist/prosetta.wasm main.go
	@if [ -f "$$(go env GOROOT)/lib/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" dist/wasm_exec.js; \
	elif [ -f "$$(go env GOROOT)/misc/wasm/wasm_exec.js" ]; then \
		cp "$$(go env GOROOT)/misc/wasm/wasm_exec.js" dist/wasm_exec.js; \
	else \
		echo "Downloading official wasm_exec.js for Go..." && \
		curl -sSL https://raw.githubusercontent.com/golang/go/master/lib/wasm/wasm_exec.js -o dist/wasm_exec.js; \
	fi
	@echo "WASM target and runtime bridge ready in dist/"

# ------------------------------------------------------------------------------
# Workspace Generation Shortcuts
# ------------------------------------------------------------------------------

# Scaffold a new category (e.g. just category path=algorithms/math)
category path="":
    @echo "Scaffolding category..."
    go run ./cmd/prosetta generate category {{path}}

# Scaffold a new concept (e.g. just concept path=algorithms/math/two-sum)
concept path="":
    @echo "Scaffolding concept..."
    go run ./cmd/prosetta generate concept {{path}}

# Scan src/ workspace directory and regenerate manifest.json
manifest:
    @echo "Generating manifest.json..."
    go run ./cmd/prosetta generate manifest

# ------------------------------------------------------------------------------
# Execution Shortcuts
# ------------------------------------------------------------------------------

# Run evaluation on public cases (e.g. just run src/fundamentals/hello-world zig)
run concept_path="" lang="":
    go run ./cmd/prosetta run {{concept_path}} {{ if lang != "" { "--lang=" + lang } else { "" } }}

# Run full evaluation against all cases (public + hidden)
submit concept_path="" lang="":
    go run ./cmd/prosetta run {{concept_path}} --all {{ if lang != "" { "--lang=" + lang } else { "" } }}

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

# Remove built binaries and test coverage files
clean:
    @echo "Cleaning build artifacts..."
    rm -rf bin/
    rm -f coverage.out coverage.html
