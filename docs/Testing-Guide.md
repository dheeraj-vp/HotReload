# Testing Guide: Hot Reload Engine

This guide provides comprehensive manual testing instructions to verify all components of Hot Reload Engine. Each test includes clear step-by-step instructions with terminal switching guidance and expected outputs.

## 📋 Important Notes Before Starting

###  Process Management Rules:
- **Background Processes**: Commands ending with `&` run in background - NO Ctrl+C needed
- **Foreground Processes**: Commands without `&` run in foreground - USE Ctrl+C to stop
- **Parallel Commands**: Multiple terminals can run commands simultaneously
- **Cleanup**: Always run `pkill -f hotreload` after each test section

###  Terminal Usage:
- **Terminal 1**: Primary terminal for starting hotreload and making changes
- **Terminal 2**: Secondary terminal for testing server responses, curl commands, etc.
- **Switch**: Instructions clearly state when to switch between terminals

##  Quick Start Test

### 1. Build the Tool
**Terminal 1:**
```bash
cd /home/dheeraj/Projects/IMP/HotReload
go build -o hotreload .
```
**Expected Output:**
```
%
```
**Why:** Verifies the Go project compiles successfully and creates the hotreload binary.

---

### 2. Test Basic CLI Commands
**Terminal 1:**
```bash
./hotreload --help
```
**Expected Output:** (Full help text with usage, flags, and examples)
```
hotreload - Automatic rebuild and restart for development

USAGE:
  hotreload --root <directory> --build <command> --exec <command>
...
```
**Why:** Confirms CLI help system works and displays proper usage information.

**Terminal 1:**
```bash
./hotreload --version
```
**Expected Output:**
```
hotreload version 1.0.0
Go version: go1.22.0
Built: 2026-03-06T10:00:00Z
```
**Why:** Verifies version information is displayed correctly.

---

### 3. Test CLI Validation
**Terminal 1:**
```bash
./hotreload --root ./nonexistent --build "go build" --exec "./server"
```
**Expected Output:**
```
2026/03/08 ERROR configuration validation failed error="root directory does not exist: /home/dheeraj/Projects/IMP/HotReload/nonexistent"
```
**Why:** Tests proper validation of invalid directory paths.

**Terminal 1:**
```bash
./hotreload --root . --build "" --exec "./server"
```
**Expected Output:**
```
2026/03/08 ERROR configuration validation failed error="required flag --build not provided"
```
**Why:** Tests validation of empty build command.

**Terminal 1:**
```bash
./hotreload --root .
```
**Expected Output:**
```
2026/03/08 ERROR configuration validation failed error="required flag --build not provided"
```
**Why:** Tests validation of missing required flags.

**Terminal 1:**
```bash
./hotreload --unknown-flag
```
**Expected Output:**
```
2026/03/08 ERROR failed to parse arguments error="unknown flag: --unknown-flag"
```
**Why:** Tests handling of unknown command-line flags.

---

## 📋 Comprehensive Testing Checklist

### ✅ File Watcher Testing

#### Test 1: Basic File Detection
**Terminal 1:** Start hotreload in background (runs automatically, no Ctrl+C needed)
```bash
./hotreload --root ./testserver --build "go build -o ./bin/testserver ." --exec "./bin/testserver PORT=8081" --verbose > watcher_test.log 2>&1 &
```
**Expected Output:**
```
[1] 12345
```
**Process Status:** Running in background - you can immediately type next commands
**Why:** Starts the hot reload tool in background to monitor file changes while allowing you to continue typing commands.

**Wait 3 seconds**, then check the log:
```bash
tail -10 watcher_test.log
```
**Expected Output:** (Initial build and server start logs)
```
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"hotreload starting","root":"..."}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"triggering initial build"}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"build starting","cmd":"go build -o ./bin/testserver ."}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"build succeeded","duration":"xxms","output":""}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"server started","pid":12345,"uptime":"x.xxxµs"}
```
**Why:** Confirms initial build triggers immediately and server starts successfully.

**Terminal 1:** Make a file change
```bash
echo "// Test change" >> ./testserver/main.go
```
**Wait 2 seconds**, then check the log:
```bash
tail -5 watcher_test.log
```
**Expected Output:** (File change detection and rebuild)
```
{"time":"2026-03-08Txx:xx:xx.xxx","level":"DEBUG","msg":"go file changed","path":".../main.go","op":1}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"build starting","cmd":"go build -o ./bin/testserver ."}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"build succeeded","duration":"xxms","output":""}
```
**Why:** Verifies file watcher detects .go file changes and triggers rebuilds.

#### Test 2: Ignore Patterns
**Terminal 1:** Create files that should be ignored
```bash
touch ./testserver/.gitignore
touch ./testserver/temp.tmp
touch ./testserver/file.swp
mkdir ./testserver/node_modules
touch ./testserver/node_modules/index.js
```
**Wait 2 seconds**, then check the log:
```bash
tail -10 watcher_test.log
```
**Expected Output:** No build should be triggered, only "ignoring event" messages
```
{"time":"2026-03-08Txx:xx:xx.xxx","level":"DEBUG","msg":"ignoring event","path":".../temp.tmp","op":1}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"DEBUG","msg":"ignoring event","path":".../file.swp","op":1}
```
**Why:** Confirms file filtering ignores non-.go files and specified patterns.

#### Test 3: Dynamic Directory Creation
**Terminal 1:** Create nested directory with Go file
```bash
mkdir -p ./testserver/newdir/nested
echo "package newdir" > ./testserver/newdir/nested/test.go
```
**Wait 2 seconds**, then check the log:
```bash
tail -10 watcher_test.log
```
**Expected Output:** Directory detection and file change
```
{"time":"2026-03-08Txx:xx:xx.xxx","level":"DEBUG","msg":"watching directory","path":".../newdir"}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"DEBUG","msg":"watching directory","path":".../newdir/nested"}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"DEBUG","msg":"added new directory to watch list","path":".../newdir"}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"DEBUG","msg":"go file changed","path":".../test.go","op":1}
```
**Why:** Verifies dynamic directory creation and automatic watching.

**Clean up:**
```bash
pkill -f hotreload
```
**Process Control:** Use this command to stop background processes - Ctrl+C is NOT needed for background processes
**Why:** Properly terminates the background hotreload process and cleans up before the next test.

---

### ✅ Debouncer Testing

#### Test 1: Rapid Changes Debounce
**Terminal 1:** Start hotreload
```bash
./hotreload --root ./testserver --build "go build -o ./bin/testserver ." --exec "./bin/testserver PORT=8082" --verbose > debounce_test.log 2>&1 &
```
**Wait 3 seconds** for initial build, then make rapid changes:
```bash
echo "// Change 1" >> ./testserver/main.go
echo "// Change 2" >> ./testserver/main.go
echo "// Change 3" >> ./testserver/main.go
```
**Wait 2 seconds**, then check the log:
```bash
tail -10 debounce_test.log
```
**Expected Output:** Only ONE build should occur after the last change
```
{"time":"2026-03-08Txx:xx:xx.xxx","level":"DEBUG","msg":"go file changed","path":".../main.go","op":1}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"DEBUG","msg":"go file changed","path":".../main.go","op":1}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"DEBUG","msg":"go file changed","path":".../main.go","op":1}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"build starting","cmd":"go build -o ./bin/testserver ."}
```
**Why:** Confirms 300ms debounce prevents multiple rapid builds.

**Clean up:**
```bash
pkill -f hotreload
```
**Process Control:** Use this command to stop background processes - Ctrl+C is NOT needed for background processes
**Why:** Properly terminates the background hotreload process and cleans up before the next test.

---

### ✅ Builder Testing

#### Test 1: Build Cancellation
**Terminal 1:** Start hotreload
```bash
./hotreload --root ./testserver --build "go build -o ./bin/testserver ." --exec "./bin/testserver PORT=8083" --verbose > cancel_test.log 2>&1 &
```
**Wait 3 seconds** for initial build, then make two rapid changes:
```bash
echo "// First change" >> ./testserver/main.go
sleep 0.1
echo "// Second change" >> ./testserver/main.go
```
**Wait 2 seconds**, then check the log:
```bash
tail -10 cancel_test.log
```
**Expected Output:** Build cancellation message
```
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"cancelling previous build"}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"build starting","cmd":"go build -o ./bin/testserver ."}
```
**Why:** Verifies concurrent builds are cancelled properly.

**Clean up:**
```bash
pkill -f hotreload
```
**Process Control:** Use this command to stop background processes - Ctrl+C is NOT needed for background processes
**Why:** Properly terminates the background hotreload process and cleans up before the next test.

---

### ✅ Process Manager Testing

#### Test 1: Graceful Shutdown
**Terminal 1:** Create graceful shutdown script
```bash
cat > ./testserver/graceful.go << 'EOF'
package main
import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Graceful server starting...")
	
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM)
		sig := <-sigChan
		fmt.Printf("Received %v, shutting down gracefully...\n", sig)
		time.Sleep(2 * time.Second)
		fmt.Println("Graceful shutdown complete")
		os.Exit(0)
	}()
	
	for i := 0; i < 30; i++ {
		fmt.Printf("Working... %d\n", i+1)
		time.Sleep(1 * time.Second)
	}
}
EOF
```

**Terminal 1:** Test graceful shutdown
```bash
./hotreload --root ./testserver --build "go build -o ./bin/graceful ./graceful.go" --exec "./bin/graceful" --verbose > graceful_test.log 2>&1 &
```
**Process Status:** Running in background - you can immediately type next commands
**Wait 3 seconds** for startup, then stop:
```bash
pkill -f hotreload
```
**Process Control:** Use pkill command for background processes - Ctrl+C is NOT needed
**Expected Output:** Graceful shutdown messages in the log file
```
Graceful server starting...
Working... 1
Working... 2
Received terminated, shutting down gracefully...
Graceful shutdown complete
```
**Why:** Verifies SIGTERM is sent and server shuts down gracefully.

#### Test 2: Force Kill for Stubborn Process
**Terminal 1:** Create stubborn process script
```bash
cat > ./testserver/stubborn.go << 'EOF'
package main
import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("Stubborn server starting...")
	
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGTERM)
		for range sigChan {
			fmt.Println("Ignoring SIGTERM, I'm stubborn!")
		}
	}()
	
	for {
		fmt.Println("Stubborn server running...")
		time.Sleep(1 * time.Second)
	}
}
EOF
```

**Terminal 1:** Test force kill
```bash
./hotreload --root ./testserver --build "go build -o ./bin/stubborn ./stubborn.go" --exec "./bin/stubborn" --verbose > stubborn_test.log 2>&1 &
```
**Process Status:** Running in background - you can immediately type next commands
**Wait 3 seconds** for startup, then stop:
```bash
pkill -f hotreload
```
**Process Control:** Use pkill command for background processes - Ctrl+C is NOT needed
**Expected Output:** Force kill message after timeout in the log file
```
Stubborn server starting...
Ignoring SIGTERM, I'm stubborn!
Stubborn server running...
...
failed to stop existing process: process did not terminate gracefully, force killed
```
**Why:** Verifies force kill after 5-second timeout for processes ignoring SIGTERM.

---

### ✅ Crash Guard Testing

#### Test 1: Crash Detection and Backoff
**Terminal 1:** Create crashing server
```bash
cat > ./testserver/crasher.go << 'EOF'
package main
import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Crashing server starting...")
	fmt.Println("About to crash!")
	os.Exit(1)
}
EOF
```

**Terminal 1:** Test crash protection
```bash
./hotreload --root ./testserver --build "go build -o ./bin/crasher ./crasher.go" --exec "./bin/crasher" --verbose > crash_test.log 2>&1 &
```
**Wait 5 seconds**, then check the log:
```bash
tail -15 crash_test.log
```
**Expected Output:** Crash detection and backoff
```
Crashing server starting...
About to crash!
{"time":"2026-03-08Txx:xx:xx.xxx","level":"WARN","msg":"server crash detected"}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"WARN","msg":"server crash recorded","crash_count":1,"backoff_delay":"1s","can_restart":true}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"waiting for backoff delay before restart","delay":"1s"}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"build starting","cmd":"go build -o ./bin/crasher ./crasher.go"}
```
**Why:** Verifies crash detection and exponential backoff restart.

**Clean up:**
```bash
pkill -f hotreload
```
**Process Control:** Use this command to stop background processes - Ctrl+C is NOT needed for background processes
**Why:** Properly terminates the background hotreload process and cleans up before the next test.

---

## 🌐 End-to-End Integration Test

### Complete Workflow Verification

**Step 1: Clean Environment**
**Terminal 1:**
```bash
rm -f ./testserver/bin/*
pkill -f hotreload
```
**Why:** Ensures clean test environment.

**Step 2: Start Hot Reload**
**Terminal 1:**
```bash
./hotreload --root ./testserver --build "go build -o ./bin/testserver ." --exec "./bin/testserver PORT=8084" --verbose > integration_test.log 2>&1 &
```
**Expected Output:**
```
[1] 12345
```
**Why:** Starts the complete hot reload system.

**Step 3: Verify Initial Build**
**Wait 3 seconds**, then check:
```bash
tail -10 integration_test.log
```
**Expected Output:** Initial build and server start
```
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"triggering initial build"}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"build succeeded","duration":"xxms"}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"server started","pid":12345}
```
**Why:** Confirms initial build triggers immediately.

**Step 4: Test Server Response**
**Terminal 2:**
```bash
curl -s http://localhost:8084 | head -5
```
**Expected Output:** HTML content from demo server
```
<!DOCTYPE html>
<html>
<head><title>Demo Server</title></head>
...
```
**Why:** Verifies server is responding correctly.

**Step 5: Test File Change Detection**
**Terminal 1:**
```bash
echo "// Integration test change" >> ./testserver/main.go
```
**Wait 2 seconds**, then check logs:
```bash
tail -5 integration_test.log
```
**Expected Output:** Rebuild triggered
```
{"time":"2026-03-08Txx:xx:xx.xxx","level":"DEBUG","msg":"go file changed","path":".../main.go"}
{"time":"2026-03-08Txx:xx:xx.xxx","level":"INFO","msg":"build starting","cmd":"go build -o ./bin/testserver ."}
```
**Why:** Confirms file change detection works.

**Step 6: Test Rapid Changes (Debounce)**
**Terminal 1:**
```bash
echo "// Rapid 1" >> ./testserver/main.go
echo "// Rapid 2" >> ./testserver/main.go
echo "// Rapid 3" >> ./testserver/main.go
```
**Wait 2 seconds**, then verify only one build occurred in logs.
**Why:** Confirms debounce logic prevents multiple builds.

**Step 7: Test Nested Directory**
**Terminal 1:**
```bash
mkdir -p ./testserver/deep/nested
echo "package deep" > ./testserver/deep/nested/helper.go
```
**Wait 2 seconds**, check logs for directory detection.
**Why:** Verifies nested directory watching.

**Step 8: Verify Performance**
**Terminal 1:**
```bash
echo "// Performance test" >> ./testserver/main.go
```
**Note the timestamp** in logs when change is made and when rebuild completes.
**Expected:** Total time < 2 seconds (typically ~600ms)
**Why:** Confirms performance requirements are met.

**Step 9: Clean Up**
```bash
pkill -f hotreload
```
**Why:** Properly shuts down the system.

---

## 🔧 Performance Testing

### Build Time Measurement
**Terminal 1:**
```bash
time go build -o ./bin/testserver ./testserver/main.go
```
**Expected Output:**
```
real    0m0.1XXs
user    0m0.0XXs
sys     0m0.0XXs
```
**Why:** Measures raw build time performance.

### Memory Usage Check
**Terminal 1:**
```bash
./hotreload --root ./testserver --build "go build -o ./bin/testserver ." --exec "./bin/testserver PORT=8085" &
HOTRELOAD_PID=$!
sleep 5
ps aux | grep $HOTRELOAD_PID
```
**Expected Output:** Memory usage < 10MB
```
dheeraj   12345  0.0  0.1  8234567  1234 pts/1    SNl  xx:xx   0:00 ./hotreload ...
```
**Why:** Verifies memory efficiency.

**Clean up:**
```bash
pkill -f hotreload
```
**Process Control:** Use this command to stop background processes - Ctrl+C is NOT needed for background processes
**Why:** Properly terminates the background hotreload process and cleans up before the next test.

---

## 🚀 Production Readiness Verification

If all tests pass with the expected outputs, the Hot Reload Engine is **PRODUCTION READY** with:

- ✅ **Reliability**: Handles all edge cases gracefully
- ✅ **Performance**: Meets all requirements (~600ms restart time)
- ✅ **Usability**: Simple CLI with proper validation
- ✅ **Robustness**: Comprehensive error handling and recovery
- ✅ **Scalability**: Efficient resource usage for long-running use

