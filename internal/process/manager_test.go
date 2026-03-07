package process

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestManager_StartAndStop(t *testing.T) {
	ctx := context.Background()
	manager := New()

	// Create a simple test script
	tempDir := t.TempDir()
	script := filepath.Join(tempDir, "test.sh")
	err := os.WriteFile(script, []byte("#!/bin/sh\necho 'test server running'\nsleep 10"), 0755)
	assert.NoError(t, err)

	// Start process
	err = manager.Start(ctx, script, tempDir)
	assert.NoError(t, err)
	assert.Equal(t, StatusRunning, manager.Status())
	assert.Greater(t, manager.PID(), 0)

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Stop process
	err = manager.Stop()
	// Note: This might force kill the process, which is expected
	if err != nil {
		assert.Contains(t, err.Error(), "force killed")
	}
	assert.Equal(t, StatusStopped, manager.Status())
}

func TestManager_StartNonExistentCommand(t *testing.T) {
	ctx := context.Background()
	manager := New()

	tempDir := t.TempDir()
	err := manager.Start(ctx, "nonexistentcommand", tempDir)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to start process")
	assert.Equal(t, StatusFailed, manager.Status())
}

func TestManager_GracefulShutdown(t *testing.T) {
	ctx := context.Background()
	manager := New()

	// Create a script that handles SIGTERM gracefully
	tempDir := t.TempDir()
	script := filepath.Join(tempDir, "graceful.sh")
	err := os.WriteFile(script, []byte(`#!/bin/sh
trap 'echo "Received SIGTERM, exiting"; exit 0' TERM
echo "Starting graceful server"
sleep 5
`), 0755)
	assert.NoError(t, err)

	// Start process
	err = manager.Start(ctx, script, tempDir)
	assert.NoError(t, err)
	assert.Equal(t, StatusRunning, manager.Status())

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	start := time.Now()
	err = manager.Stop()
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.Equal(t, StatusStopped, manager.Status())
	assert.Less(t, duration, 2*time.Second, "Should shutdown gracefully quickly")
}

func TestManager_ForceKill(t *testing.T) {
	ctx := context.Background()
	manager := New()

	// Create a script that ignores SIGTERM
	tempDir := t.TempDir()
	script := filepath.Join(tempDir, "stubborn.sh")
	err := os.WriteFile(script, []byte(`#!/bin/sh
trap '' TERM
echo "Starting stubborn server"
sleep 30
`), 0755)
	assert.NoError(t, err)

	// Start process
	err = manager.Start(ctx, script, tempDir)
	assert.NoError(t, err)
	assert.Equal(t, StatusRunning, manager.Status())

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	start := time.Now()
	err = manager.Stop()
	duration := time.Since(start)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "force killed")
	assert.Equal(t, StatusStopped, manager.Status())
	assert.Greater(t, duration, 5*time.Second, "Should wait for timeout before force kill")
	assert.Less(t, duration, 7*time.Second, "Should not take too long")
}

func TestManager_Restart(t *testing.T) {
	ctx := context.Background()
	manager := New()

	// Create a simple test script
	tempDir := t.TempDir()
	script := filepath.Join(tempDir, "restart.sh")
	err := os.WriteFile(script, []byte("#!/bin/sh\necho 'server running'\nsleep 10"), 0755)
	assert.NoError(t, err)

	// Start first process
	err = manager.Start(ctx, script, tempDir)
	assert.NoError(t, err)
	firstPID := manager.PID()
	assert.Equal(t, StatusRunning, manager.Status())

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Restart (should stop old and start new)
	err = manager.Start(ctx, script, tempDir)
	// Note: This might force kill the old process, which is expected
	assert.NoError(t, err)
	secondPID := manager.PID()
	assert.Equal(t, StatusRunning, manager.Status())

	// PIDs should be different
	assert.NotEqual(t, firstPID, secondPID)

	// Stop
	err = manager.Stop()
	// Note: This might force kill the process, which is expected
	if err != nil {
		assert.Contains(t, err.Error(), "force killed")
	}
	assert.Equal(t, StatusStopped, manager.Status())
}

func TestManager_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	manager := New()

	// Create a long-running script
	tempDir := t.TempDir()
	script := filepath.Join(tempDir, "long.sh")
	err := os.WriteFile(script, []byte("#!/bin/sh\necho 'long server'\nsleep 30"), 0755)
	assert.NoError(t, err)

	// Start process
	err = manager.Start(ctx, script, tempDir)
	assert.NoError(t, err)
	assert.Equal(t, StatusRunning, manager.Status())

	// Give it a moment to start
	time.Sleep(100 * time.Millisecond)

	// Cancel context
	cancel()

	// Give time for cleanup
	time.Sleep(100 * time.Millisecond)

	// Process should be stopped
	assert.Equal(t, StatusStopped, manager.Status())
}

func TestManager_Uptime(t *testing.T) {
	ctx := context.Background()
	manager := New()

	// Initially no uptime
	assert.Equal(t, time.Duration(0), manager.Uptime())

	// Create a simple test script
	tempDir := t.TempDir()
	script := filepath.Join(tempDir, "uptime.sh")
	err := os.WriteFile(script, []byte("#!/bin/sh\necho 'uptime test'\nsleep 2"), 0755)
	assert.NoError(t, err)

	// Start process
	err = manager.Start(ctx, script, tempDir)
	assert.NoError(t, err)

	// Give it time to run
	time.Sleep(100 * time.Millisecond)

	// Should have uptime
	uptime := manager.Uptime()
	assert.Greater(t, uptime, 50*time.Millisecond)
	assert.Less(t, uptime, 200*time.Millisecond)

	// Stop process
	err = manager.Stop()
	assert.NoError(t, err)

	// Uptime should be 0 when stopped
	assert.Equal(t, time.Duration(0), manager.Uptime())
}

func TestManager_IsRunning(t *testing.T) {
	ctx := context.Background()
	manager := New()

	// Initially not running
	assert.False(t, manager.IsRunning())

	// Create a simple test script
	tempDir := t.TempDir()
	script := filepath.Join(tempDir, "running.sh")
	err := os.WriteFile(script, []byte("#!/bin/sh\necho 'running test'\nsleep 1"), 0755)
	assert.NoError(t, err)

	// Start process
	err = manager.Start(ctx, script, tempDir)
	assert.NoError(t, err)

	// Should be running
	assert.True(t, manager.IsRunning())

	// Stop process
	err = manager.Stop()
	// Note: This might force kill the process, which is expected
	if err != nil {
		assert.Contains(t, err.Error(), "force killed")
	}

	// Should not be running
	assert.False(t, manager.IsRunning())
}

func TestManager_CommandParsing(t *testing.T) {
	manager := New()

	tests := []struct {
		command   string
		program   string
		args      []string
		expectErr bool
	}{
		{
			command:   "./server",
			program:   "./server",
			args:      []string{},
			expectErr: false,
		},
		{
			command:   "go run main.go",
			program:   "go",
			args:      []string{"run", "main.go"},
			expectErr: false,
		},
		{
			command:   `./server --port "8080" --host "localhost"`,
			program:   "./server",
			args:      []string{"--port", "8080", "--host", "localhost"},
			expectErr: false,
		},
		{
			command:   "",
			program:   "",
			args:      nil,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			program, args, err := manager.parseCommand(tt.command)
			
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.program, program)
				assert.Equal(t, tt.args, args)
			}
		})
	}
}

func TestManager_StatusString(t *testing.T) {
	tests := []struct {
		status   Status
		expected string
	}{
		{StatusStopped, "stopped"},
		{StatusStarting, "starting"},
		{StatusRunning, "running"},
		{StatusStopping, "stopping"},
		{StatusFailed, "failed"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.status.String())
		})
	}
}
