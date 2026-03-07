.PHONY: build demo test clean help

# Build the hotreload binary
build:
	@echo "🔨 Building hotreload..."
	go build -o hotreload ./main.go
	@echo "✅ Build complete: ./hotreload"

# Run the demo server with hot reload
demo:
	@echo "🚀 Starting hot reload demo..."
	@echo "Open http://localhost:8080 to see the server"
	@echo " Edit testserver/main.go to see hot reload in action!"
	@if [ ! -f hotreload ]; then $(MAKE) build; fi
	./hotreload --root ./testserver --build "go build -o ./testserver ./testserver/main.go" --exec "./testserver" --verbose

# Run tests
test:
	@echo "Running tests..."
	go test ./... -v -cover

# Clean build artifacts
clean:
	@echo "🧹 Cleaning up..."
	rm -f hotreload
	rm -f testserver
	go clean -testcache
	@echo "✅ Clean complete"

# Show help
help:
	@echo "Hot Reload Engine - Build Targets"
	@echo ""
	@echo "build    - Build the hotreload binary"
	@echo "demo     - Run the demo server with hot reload"
	@echo "test     - Run all tests with coverage"
	@echo "clean    - Remove build artifacts"
	@echo "help     - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make demo           # Run demo"
	@echo "  make build && ./hotreload --help  # Build and see options"
