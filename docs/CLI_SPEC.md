# CLI Specification: hotreload

## Command Synopsis

```bash
hotreload --root <directory> --build <command> --exec <command> [options]
```

## Required Flags

### `--root <directory>`

Specifies the project directory to watch for file changes.

**Type**: String (file path)

**Validation**:
- Must be an existing directory
- Automatically converted to absolute path
- Symbolic links are NOT followed

**Examples**:
```bash
--root ./myproject          # Relative path (converted to absolute)
--root /home/user/project   # Absolute path
--root .                    # Current directory
```

**Errors**:
```bash
# Directory doesn't exist
$ hotreload --root /nonexistent --build "go build" --exec "./server"
ERROR root directory does not exist: /nonexistent

# Not a directory (it's a file)
$ hotreload --root main.go --build "go build" --exec "./server"
ERROR root must be a directory, got file: main.go
```

---

### `--build <command>`

Command to execute when files change. This command builds your project.

**Type**: String (shell command)

**Validation**:
- Must be non-empty
- Will be parsed into program and arguments
- Executed in the `--root` directory

**Command Parsing**:
```bash
# Simple command
--build "go build"
# Parsed as: ["go", "build"]

# Command with arguments
--build "go build -o ./bin/server ./cmd/server"
# Parsed as: ["go", "build", "-o", "./bin/server", "./cmd/server"]

# Command with flags
--build "go build -ldflags='-s -w' -o server"
# Parsed as: ["go", "build", "-ldflags='-s -w'", "-o", "server"]
```

**Working Directory**: Set to `--root` directory

**Timeout**: 60 seconds (build fails if exceeds)

**Output Capture**: Both stdout and stderr are captured and logged

**Examples**:
```bash
# Go project
--build "go build -o ./bin/server ./main.go"

# Make-based build
--build "make build"

# Custom script
--build "./build.sh"

# Multi-step (using shell)
--build "sh -c 'make clean && make build'"
```

**Errors**:
```bash
# Empty command
$ hotreload --root . --build "" --exec "./server"
ERROR build command cannot be empty

# Command not found
$ hotreload --root . --build "nonexistent-compiler" --exec "./server"
ERROR build failed: exec: "nonexistent-compiler": executable file not found in $PATH
SUGGESTION: Install the required build tool or check PATH environment variable
```

---

### `--exec <command>`

Command to execute after successful build. This runs your server/application.

**Type**: String (shell command)

**Validation**:
- Must be non-empty
- Will be parsed into program and arguments
- Executed in the `--root` directory

**Process Management**:
- Creates new process group (to kill all children)
- Logs streamed to stdout in real-time
- Graceful shutdown: SIGTERM (5s timeout)
- Forced shutdown: SIGKILL to entire process group

**Working Directory**: Set to `--root` directory

**Environment**: Inherits parent's environment variables

**Examples**:
```bash
# Run binary directly
--exec "./bin/server"

# Run with arguments
--exec "./bin/server --port 8080 --env dev"

# Run Go directly (development)
--exec "go run main.go"

# Run with environment variables
--exec "PORT=8080 ./bin/server"

# Run shell script
--exec "./start.sh"
```

**Errors**:
```bash
# Empty command
$ hotreload --root . --build "go build" --exec ""
ERROR exec command cannot be empty

# Binary not found
$ hotreload --root . --build "go build -o server" --exec "./nonexistent"
ERROR failed to start process: exec: "./nonexistent": no such file or directory
SUGGESTION: Check that build command created the expected binary

# Permission denied
$ hotreload --root . --build "go build" --exec "./server"
ERROR failed to start process: exec: "./server": permission denied
SUGGESTION: Make binary executable with: chmod +x ./server
```

---

## Optional Flags

### `--help` or `-h`

Displays usage information and exits.

**Example**:
```bash
$ hotreload --help
hotreload - Automatic rebuild and restart for development

USAGE:
  hotreload --root <directory> --build <command> --exec <command>

REQUIRED FLAGS:
  --root <directory>   Project directory to watch for changes
  --build <command>    Command to build the project
  --exec <command>     Command to run the built application

OPTIONAL FLAGS:
  --help, -h           Show this help message
  --version, -v        Show version information
  --verbose            Enable verbose logging

EXAMPLES:
  # Go HTTP server
  hotreload --root ./myproject \
            --build "go build -o ./bin/server ./cmd/server" \
            --exec "./bin/server"

  # Node.js with TypeScript
  hotreload --root ./myapp \
            --build "npm run build" \
            --exec "node dist/index.js"

For more information: https://github.com/yourorg/hotreload
```

---

### `--version` or `-v`

Displays version information and exits.

**Example**:
```bash
$ hotreload --version
hotreload version 1.0.0
Go version: go1.22.0
Built: 2026-03-06T10:00:00Z
```

---

### `--verbose`

Enables verbose logging (DEBUG level).

**Default**: INFO level

**Example**:
```bash
$ hotreload --root . --build "go build" --exec "./server" --verbose
2026-03-06T10:20:00.000Z DEBUG watcher adding directory path=/home/user/project/internal
2026-03-06T10:20:00.001Z DEBUG watcher adding directory path=/home/user/project/cmd
2026-03-06T10:20:00.100Z INFO watcher initialized paths_watched=15
...
```

---

## Exit Codes

| Code | Meaning | Example Scenario |
|------|---------|------------------|
| 0 | Success | Normal shutdown (SIGINT/SIGTERM) |
| 1 | Invalid arguments | Missing required flag |
| 2 | Initialization error | Cannot create file watcher |
| 3 | Crash limit exceeded | Server crashed 5 times in 60 seconds |
| 130 | Interrupted (SIGINT) | User pressed Ctrl+C |

**Examples**:
```bash
$ hotreload --root /invalid
echo $?
# 2

$ hotreload --root . --build "go build" --exec "./crasher"
# [After 5 rapid crashes]
echo $?
# 3
```

---

## Environment Variables

### Inherited Variables

hotreload inherits the parent shell's environment and passes it to build and exec commands.

**Commonly Used**:
- `PATH`: For finding build tools
- `GOPATH`: For Go projects
- `HOME`: User home directory
- Custom variables: Any you set

**Example**:
```bash
$ export BUILD_ENV=development
$ hotreload --root . --build "make build" --exec "./server"
# Both make and ./server see BUILD_ENV=development
```

### NO Custom Environment Variables (Current Version)

hotreload does not currently support setting custom environment variables via CLI.

**Workaround**:
```bash
# Set in shell before running hotreload
export PORT=8080
hotreload --root . --build "go build" --exec "./server"

# Or wrap in script
#!/bin/bash
export PORT=8080
exec hotreload --root . --build "go build" --exec "./server"
```

---

## Signal Handling

### SIGINT (Ctrl+C)

**Behavior**:
1. Stop watching for file changes
2. Stop running server process (graceful)
3. Wait for clean shutdown (max 5s)
4. Exit with code 130

**Example**:
```bash
$ hotreload --root . --build "go build" --exec "./server"
# [Press Ctrl+C]
^C
2026-03-06T10:25:00.000Z INFO shutting down reason="received interrupt signal"
2026-03-06T10:25:00.100Z INFO process stopping pid=12345
2026-03-06T10:25:00.300Z INFO process stopped duration=200ms
2026-03-06T10:25:00.301Z INFO shutdown complete
```

---

### SIGTERM

**Behavior**: Same as SIGINT (graceful shutdown)

**Use Case**: System shutdown, process managers

**Example**:
```bash
$ hotreload --root . --build "go build" --exec "./server" &
PID=$!
$ kill $PID
2026-03-06T10:25:00.000Z INFO shutting down reason="received termination signal"
...
```

---

## Complete Usage Examples

### Example 1: Simple Go HTTP Server

```bash
$ hotreload \
  --root ./myserver \
  --build "go build -o ./bin/server ./main.go" \
  --exec "./bin/server"

# Expected Output:
2026-03-06T10:20:00.000Z INFO hotreload starting root=/home/user/myserver
2026-03-06T10:20:00.100Z INFO watcher initialized paths_watched=5
2026-03-06T10:20:00.150Z INFO initial build starting
2026-03-06T10:20:01.543Z INFO build succeeded duration=1.543s
2026-03-06T10:20:01.544Z INFO process starting command="./bin/server"
2026-03-06T10:20:01.644Z INFO process started pid=12345
{"time":"2026-03-06T10:20:01.645Z","level":"INFO","msg":"server listening","addr":":8080"}
```

### Example 2: Go Project with Build Flags

```bash
$ hotreload \
  --root ./myproject \
  --build "go build -ldflags='-s -w' -o ./bin/server ./cmd/server" \
  --exec "./bin/server --port 8080 --env development"
```

### Example 3: Makefile-Based Build

```bash
$ hotreload \
  --root ./myapp \
  --build "make build" \
  --exec "./bin/app"
```

### Example 4: Multi-Package Go Project

```bash
$ hotreload \
  --root ./bigproject \
  --build "go build -o ./bin/api ./cmd/api" \
  --exec "./bin/api"

# Project structure:
# bigproject/
# ├── cmd/
# │   └── api/
# │       └── main.go
# ├── internal/
# │   ├── handlers/
# │   ├── models/
# │   └── db/
# └── pkg/
#     └── utils/
#
# Watches all directories: cmd/, internal/, pkg/
```

### Example 5: Verbose Debugging

```bash
$ hotreload \
  --root ./myserver \
  --build "go build -o ./bin/server ./main.go" \
  --exec "./bin/server" \
  --verbose

# Shows detailed logs:
2026-03-06T10:20:00.000Z DEBUG cli config validated root="/home/user/myserver"
2026-03-06T10:20:00.100Z DEBUG watcher adding directory path="/home/user/myserver"
2026-03-06T10:20:00.101Z DEBUG watcher ignoring path="/home/user/myserver/.git"
2026-03-06T10:20:00.150Z DEBUG file change detected path="/home/user/myserver/main.go" op="write"
...
```

---

## Error Messages Reference

### Configuration Errors

```bash
# Missing required flag
ERROR required flag --root not provided
USAGE: hotreload --root <directory> --build <command> --exec <command>

# Invalid root directory
ERROR root directory does not exist: /path/to/nonexistent
SUGGESTION: Create directory or fix path

# Root is a file, not directory
ERROR root must be a directory, got file: main.go
SUGGESTION: Specify directory containing your project
```

### Build Errors

```bash
# Build command failed
ERROR build failed duration=1.234s exit_code=1
Build output:
./main.go:10:2: syntax error: unexpected newline
SUGGESTION: Fix syntax errors and save file to trigger rebuild

# Build timeout
ERROR build timeout after 60s
SUGGESTION: Build is taking too long. Check for infinite loops or optimize build process

# Build command not found
ERROR build failed: exec: "nonexistent": executable file not found in $PATH
SUGGESTION: Install required build tool or check PATH environment variable
```

### Process Errors

```bash
# Process failed to start
ERROR failed to start process: exec: "./server": no such file or directory
SUGGESTION: Check that build command created the expected binary at ./server

# Process permission denied
ERROR failed to start process: exec: "./server": permission denied
SUGGESTION: Make binary executable with: chmod +x ./server

# Process crashed
ERROR process crashed duration=2.345s exit_code=1
WARN crash detected count=1 backoff=1s
INFO restarting in 1 second...

# Crash limit exceeded
ERROR crash limit exceeded crashes=5 window=60s
ERROR auto-restart disabled
SUGGESTIONS:
  - Check server logs above for error messages
  - Fix code errors causing crashes
  - Verify build command is correct
  - Check for port conflicts or missing dependencies
```

### Watch Errors

```bash
# Too many files to watch
WARN file descriptor limit approaching current=950 limit=1024
SUGGESTION: Increase ulimit -n or add more ignore patterns

# Permission denied on directory
ERROR cannot watch directory path=/restricted error="permission denied"
SUGGESTION: Check directory permissions or run with appropriate privileges

# Filesystem event error
ERROR watcher error error="no space left on device"
SUGGESTION: Free up disk space or inode space
```

---

## Performance Characteristics

### Latency Targets

| Operation | Target | Acceptable | Failure |
|-----------|--------|------------|---------|
| File save → rebuild start | < 400ms | < 600ms | > 1s |
| Build duration (simple) | < 1.5s | < 2.5s | > 5s |
| Process restart | < 500ms | < 1s | > 2s |
| **Total: Save → Running** | **< 2s** | **< 3s** | **> 5s** |

### Resource Usage

| Resource | Idle | Active | Limit |
|----------|------|--------|-------|
| Memory | < 20MB | < 50MB | 100MB |
| CPU | < 1% | < 50% | - |
| File Descriptors | ~100 | ~500 | 1024 |

---

## Compatibility

### Operating Systems

| OS | Status | Notes |
|----|--------|-------|
| Linux | ✅ Supported | Tested on Ubuntu 24.04, Fedora 39 |
| macOS | ✅ Supported | Tested on macOS 13+ |
| Windows | ⚠️ Partial | Process group handling differs, may have issues |

### Go Versions

- **Minimum**: Go 1.21
- **Recommended**: Go 1.22+
- **Tested**: Go 1.22.0

### Build Tools

Works with any build tool that:
- Exits with code 0 on success
- Exits with non-zero on failure
- Writes errors to stdout/stderr

**Tested Build Tools**:
- `go build` ✅
- `make` ✅
- `npm run build` ✅
- `cargo build` ✅
- Custom shell scripts ✅

---

## Frequently Asked Questions

### Q: Can I watch multiple directories?

**A**: Not currently. Specify the root directory that contains all code you want to watch.

**Workaround**: Use the highest common ancestor directory.

```bash
# Instead of watching both:
# /project/frontend
# /project/backend

# Watch parent:
--root /project
```

### Q: Can I exclude certain subdirectories?

**A**: Certain patterns are automatically excluded (see File Filtering). Custom exclude patterns are not yet supported.

**Current Auto-Excludes**:
- `.git/`
- `node_modules/`
- `*.swp`, `*.tmp`
- `.DS_Store`
- Hidden files (starting with `.`)

### Q: Can I run multiple servers simultaneously?

**A**: Not in a single hotreload instance. Each instance manages one server.

**Workaround**: Run multiple hotreload instances in separate terminals.

```bash
# Terminal 1: API server
hotreload --root ./api --build "..." --exec "./bin/api"

# Terminal 2: Worker server
hotreload --root ./worker --build "..." --exec "./bin/worker"
```

### Q: Does it work with Docker?

**A**: Not directly. hotreload is designed for local development.

**Alternative**: Use Docker's built-in watch mode or mount volumes.

### Q: Can I disable auto-restart after crash?

**A**: Auto-restart stops after 5 crashes in 60 seconds. This is not configurable.

**Rationale**: Prevents infinite crash loops that consume resources.

### Q: How do I debug why my server keeps crashing?

**A**: Check the server logs printed to stdout. Enable `--verbose` for detailed information.

```bash
hotreload --root . --build "go build" --exec "./server" --verbose
```

---

## Tips and Best Practices

### 1. Use Absolute Paths in Build/Exec Commands

```bash
# Good (relative to --root)
--build "go build -o ./bin/server ./main.go"
--exec "./bin/server"

# Avoid (absolute paths may not work across systems)
--build "go build -o /home/user/project/bin/server"
```

### 2. Build Output to Dedicated Directory

```bash
# Good (clean separation)
--build "go build -o ./bin/server"
--exec "./bin/server"

# Avoid (pollutes source directory)
--build "go build -o ./server"
```

### 3. Use Make for Complex Builds

```makefile
# Makefile
.PHONY: build
build:
	go generate ./...
	go build -ldflags="-s -w" -o ./bin/server ./cmd/server

.PHONY: clean
clean:
	rm -rf ./bin
```

```bash
hotreload --root . --build "make build" --exec "./bin/server"
```

### 4. Graceful Shutdown in Your Server

Ensure your server handles SIGTERM for clean shutdown:

```go
// Example Go server
func main() {
    server := &http.Server{Addr: ":8080"}
    
    go func() {
        if err := server.ListenAndServe(); err != http.ErrServerClosed {
            log.Fatal(err)
        }
    }()
    
    // Wait for SIGTERM
    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGTERM, syscall.SIGINT)
    <-sig
    
    // Graceful shutdown (5s timeout)
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    server.Shutdown(ctx)
}
```

### 5. Development vs Production Builds

```bash
# Development (fast build, with debug info)
hotreload --root . \
  --build "go build -o ./bin/server" \
  --exec "./bin/server"

# Production-like (optimized, stripped)
hotreload --root . \
  --build "go build -ldflags='-s -w' -o ./bin/server" \
  --exec "./bin/server"
```

---

This CLI specification provides complete documentation for using hotreload. All behavior is deterministic and testable.
