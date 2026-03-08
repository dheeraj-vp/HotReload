.PHONY: build test clean run-demo install lint test-coverage help fmt vet deps tidy ci check profile-cpu profile-mem dev watch

# Build configuration
BINARY_NAME=hotreload
TESTSERVER_BINARY=testserver/bin/server
GO=go
GOFLAGS=-ldflags="-s -w"

# Default target
.DEFAULT_GOAL := help

## build: Build hotreload binary
build:
	@echo "Building $(BINARY_NAME)..."
	@$(GO) build $(GOFLAGS) -o $(BINARY_NAME) main.go
	@echo "Build complete: ./$(BINARY_NAME)"

## test: Run all tests with race detector and coverage
test:
	@echo "Running tests with coverage..."
	@$(GO) test -v -race -cover ./...
	@echo "Tests complete"

## test-coverage: Run tests with coverage report
test-coverage:
	@echo "Running tests with coverage report..."
	@$(GO) test -race -coverprofile=coverage.out ./...
	@$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"
	@$(GO) tool cover -func=coverage.out | grep total

## test-short: Run fast tests only (skip integration tests)
test-short:
	@echo "Running short tests..."
	@$(GO) test -short -v ./...
	@echo "Short tests complete"

## test-integration: Run integration tests only
test-integration:
	@echo "Running integration tests..."
	@$(GO) test -v -run Integration ./...
	@echo "Integration tests complete"

## bench: Run benchmarks
bench:
	@echo "Running benchmarks..."
	@$(GO) test -bench=. -benchmem ./...

## build-testserver: Build the test server
build-testserver:
	@echo "Building test server..."
	@mkdir -p testserver/bin
	@cd testserver && $(GO) build -o bin/server main.go handlers.go
	@echo "Test server built: $(TESTSERVER_BINARY)"

## run-demo: Build and run hot reload demo with test server
run-demo: build build-testserver
	@echo "Starting hot reload demo..."
	@echo "Edit testserver/handlers.go to see hot reload in action"
	@echo "Open http://localhost:8080 to see the server"
	@echo "Press Ctrl+C to stop"
	@echo ""
	@./$(BINARY_NAME) \
		--root ./testserver \
		--build "go build -o ./bin/server ./main.go ./handlers.go" \
		--exec "./bin/server" \
		--verbose

## clean: Remove all build artifacts and temporary files
clean:
	@echo "Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@rm -rf testserver/bin
	@rm -f coverage.out coverage.html
	@rm -f *.test
	@rm -f cpu.prof mem.prof
	@$(GO) clean -testcache
	@echo "Clean complete"

## install: Install hotreload to GOPATH/bin
install:
	@echo "Installing $(BINARY_NAME)..."
	@$(GO) install $(GOFLAGS)
	@echo "Installed to $$(go env GOPATH)/bin/$(BINARY_NAME)"

## lint: Run golangci-lint (requires golangci-lint installed)
lint:
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
		echo "Lint complete"; \
	else \
		echo "golangci-lint not installed. Install with:"; \
		echo "  curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin"; \
	fi

## fmt: Format all Go files
fmt:
	@echo "Formatting code..."
	@$(GO) fmt ./...
	@echo "Format complete"

## vet: Run go vet
vet:
	@echo "Running go vet..."
	@$(GO) vet ./...
	@echo "Vet complete"

## deps: Download and verify dependencies
deps:
	@echo "Downloading dependencies..."
	@$(GO) mod download
	@$(GO) mod verify
	@echo "Dependencies verified"

## tidy: Tidy go.mod and go.sum
tidy:
	@echo "Tidying dependencies..."
	@$(GO) mod tidy
	@echo "Dependencies tidied"

## check: Run all quality checks (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "All checks passed"

## ci: CI pipeline (runs in CI/CD environments)
ci: deps fmt vet test
	@echo "CI checks complete"

## profile-cpu: Run CPU profiling
profile-cpu:
	@echo "Running CPU profile..."
	@$(GO) test -cpuprofile=cpu.prof -bench=. ./...
	@echo "View profile with: go tool pprof cpu.prof"

## profile-mem: Run memory profiling
profile-mem:
	@echo "Running memory profile..."
	@$(GO) test -memprofile=mem.prof -bench=. ./...
	@echo "View profile with: go tool pprof mem.prof"

## dev: Run hot reload on itself (meta development!)
dev: build
	@echo "Running hot reload on itself..."
	@./$(BINARY_NAME) \
		--root . \
		--build "go build -o $(BINARY_NAME) main.go" \
		--exec "./$(BINARY_NAME) --help"

## watch: Alias for run-demo
watch: run-demo

## help: Show comprehensive help message
help:
	@echo "Hot Reload Engine - Professional Development Automation"
	@echo ""
	@echo "Build Targets:"
	@echo "  build          - Build hotreload binary"
	@echo "  build-testserver - Build test server for demo"
	@echo "  install        - Install to GOPATH/bin"
	@echo ""
	@echo "Test Targets:"
	@echo "  test           - Run all tests with coverage"
	@echo "  test-coverage  - Generate HTML coverage report"
	@echo "  test-short     - Run fast tests only"
	@echo "  test-integration - Run integration tests only"
	@echo "  bench          - Run performance benchmarks"
	@echo ""
	@echo "Quality Targets:"
	@echo "  fmt            - Format all Go files"
	@echo "  vet            - Run go vet analysis"
	@echo "  lint           - Run golangci-lint"
	@echo "  check          - Run all quality checks"
	@echo ""
	@echo "Demo Targets:"
	@echo "  run-demo       - Build and run demo with hot reload"
	@echo "  watch          - Alias for run-demo"
	@echo "  dev            - Run hot reload on itself"
	@echo ""
	@echo "Utility Targets:"
	@echo "  clean          - Remove all build artifacts"
	@echo "  deps           - Download and verify dependencies"
	@echo "  tidy           - Tidy go.mod and go.sum"
	@echo ""
	@echo "Profiling Targets:"
	@echo "  profile-cpu    - Run CPU profiling"
	@echo "  profile-mem    - Run memory profiling"
	@echo ""
	@echo "CI/CD Targets:"
	@echo "  ci             - Full CI pipeline"
	@echo ""
	@echo "Examples:"
	@echo "  make run-demo              # Quick demo"
	@echo "  make build && ./hotreload    # Build and test"
	@echo "  make test-coverage && open coverage.html  # View coverage"
	@echo "  make check                 # Full quality check"
	@echo ""
	@echo "For documentation: see docs/README.md"
