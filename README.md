# HotReload

Production-grade automatic rebuild and restart tool for Go development. Watch your project, rebuild on changes, and restart your server - all automatically with enterprise-level reliability.

---

## ✨ Features

-  **Auto-rebuild**: Detects file changes and rebuilds your project
-  **Auto-restart**: Restarts your server after successful builds
-  **Smart debouncing**: Handles rapid file changes elegantly (300ms delay)
-  **Process management**: Graceful shutdown with forceful fallback
-  **Crash protection**: Prevents rapid restart loops with exponential backoff
-  **Real-time logs**: Streams server output immediately
-  **Recursive watching**: Monitors nested directories dynamically
-  **Production ready**: Comprehensive testing and error handling

---

## 🚀 Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/dheeraj-vp/HotReload.git
cd HotReload

# Build from source
make build

# Or run directly
go run . --root ./myproject --build "go build -o ./bin/server" --exec "./bin/server"
```

### Basic Usage

```bash
hotreload --root ./myproject \
          --build "go build -o ./bin/server ./cmd/server" \
          --exec "./bin/server"
```

That's it! Edit your code and watch it rebuild and restart automatically.

### Quick Demo

```bash
# Run the included demo server
make run-demo

# Open http://localhost:8080 to see the server
# Edit testserver/handlers.go to see hot reload in action
```

---

## 📋 Requirements

- **Go 1.21+** 
- **Linux or macOS** (Windows partially supported)

---

## 🏗️ Architecture

HotReload implements a sophisticated event-driven architecture with six core components working in concert:

<p align="center">
<img src="docs/assets/System-Architecture.png" width="500">
</p>

**Key Technical Features:**
- **Event-driven communication** using Go channels (CSP pattern)
- **Separation of concerns** with modular component design
- **Fault tolerance** with comprehensive error handling
- **Performance optimization** with sub-600ms restart times
- **Resource efficiency** with 8-15MB typical memory usage

### Data Flow

The complete hot reload workflow from file change to server restart:

<p align="center">
<img src="docs/assets/Dataflow-Happy.png" width="800">
</p>

📖 **For detailed architecture, component interfaces, and design decisions, see [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)**

---

## 💻 Command Line Interface

### Required Flags
```bash
--root <directory>     # Project directory to watch
--build <command>      # Build command to execute
--exec <command>       # Server command to run
```

### Optional Flags
```bash
--help, -h            # Show usage information
--version, -v         # Show version information
--verbose             # Enable debug logging
```

### Example Usage
```bash
# Go HTTP server
hotreload --root ./api \
          --build "go build -o ./bin/server ./cmd/server" \
          --exec "./bin/server --port 8080"

# Makefile-based build
hotreload --root ./myapp \
          --build "make build" \
          --exec "./bin/app"

# Development with verbose logging
hotreload --root . \
          --build "go build -o ./server" \
          --exec "./server" \
          --verbose
```

📖 **For complete CLI specification, validation rules, and error handling, see [docs/CLI_SPEC.md](./docs/CLI_SPEC.md)**

---

## ⚡ Performance

### Benchmarks
- **Restart Time**: ~600ms typical (target: <2s)
- **Memory Usage**: 8-15MB (depends on project size)
- **CPU Usage**: <1% idle, <20% during builds
- **File Latency**: <50ms from change to detection

### Resource Efficiency
- **File Descriptors**: 1 per watched directory
- **Goroutines**: 4-6 total (main + components)
- **Build Cancellation**: <100ms to stop in-progress builds

---

## 🧪 Testing

HotReload includes comprehensive testing covering all components and edge cases:

### Test Coverage
- **Unit Tests**: All components with >85% coverage
- **Integration Tests**: End-to-end workflow validation
- **Performance Tests**: Sub-2 second restart verification
- **Stress Tests**: Rapid file changes and crash loops

### Manual Testing
Complete step-by-step testing guide for verification:
- CLI functionality and validation
- File watcher behavior and ignore patterns
- Debouncer timing and rapid changes
- Builder success/failure scenarios
- Process management and crash recovery

📖 **For comprehensive testing procedures and verification checklist, see [docs/Testing-Guide.md](./docs/Testing-Guide.md)**

---

## 🛡️ Reliability Features

### Error Handling
- **Build Failures**: Graceful handling, no restart on failure
- **Process Crashes**: Automatic detection with exponential backoff
- **Stubborn Processes**: Force kill after 5-second timeout
- **Resource Exhaustion**: Proper cleanup and error reporting

### Crash Protection
- **Exponential Backoff**: 1s, 2s, 4s, 8s, 16s, 30s delays
- **Crash Limits**: 5 crashes in 60 seconds triggers stop
- **Automatic Recovery**: Resets counter after 60 seconds of stability

### Edge Cases
- **Rapid File Changes**: Debounced to single rebuild
- **Directory Creation**: Auto-added to watch list
- **Permission Issues**: Clear error messages with suggestions
- **Network Issues**: No external dependencies

---

## 🔧 Advanced Usage

### Process Management
```bash
# Graceful shutdown (SIGTERM) - 5 second timeout
# Force shutdown (SIGKILL) - immediate termination
# Process groups - kills all child processes
```

### Environment Variables
```bash
# Inherit parent environment
export PORT=8080
export DEBUG=true
hotreload --root . --build "make build" --exec "./bin/server"
```

### Build Optimization
```bash
# Development build (fast)
--build "go build -o ./bin/server"

# Production build (optimized)
--build "go build -ldflags='-s -w' -o ./bin/server"
```

---

## 📚 Documentation

- **[docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)** - Complete system design, component interfaces, and technical architecture
- **[docs/CLI_SPEC.md](./docs/CLI_SPEC.md)** - Comprehensive command-line interface specification with all options
- **[docs/Testing-Guide.md](./docs/Testing-Guide.md)** - Step-by-step testing procedures and verification guide

---

## 🚀 Production Deployment

### Binary Distribution
```bash
# Clone and build
git clone https://github.com/dheeraj-vp/HotReload.git
cd HotReload
make build

# Static binary for deployment
CGO_ENABLED=0 go build -ldflags="-s -w" -o hotreload .

# Cross-platform builds
GOOS=linux GOARCH=amd64 go build -o hotreload-linux .
GOOS=darwin GOARCH=amd64 go build -o hotreload-macos .
```

### Build Automation with Makefile
The project includes a comprehensive Makefile for all development tasks:

```bash
# View all available targets
make help

# Common development tasks
make build          # Build hotreload binary
make run-demo       # Quick demo with test server
make test          # Run all tests with coverage
make check          # Run quality checks (fmt, vet, lint, test)
make clean          # Remove build artifacts

# Advanced features
make test-coverage  # Generate HTML coverage report
make profile-cpu    # CPU profiling
make ci            # Full CI pipeline
```

### Best Practices
- Use absolute paths in build/exec commands
- Build to dedicated `./bin/` directory
- Implement graceful shutdown in your server
- Monitor logs with `--verbose` for debugging

---

## 🔗 Links

- **GitHub Repository**: https://github.com/dheeraj-vp/HotReload
- **Architecture Documentation**: [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md)
- **CLI Reference**: [docs/CLI_SPEC.md](./docs/CLI_SPEC.md)
- **Testing Guide**: [docs/Testing-Guide.md](./docs/Testing-Guide.md)
- **Build System**: See `make help` for all available targets

---

**HotReload: Enterprise-grade development automation for Go projects** 🚀
