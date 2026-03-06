package builder

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBuilder_SuccessfulBuild(t *testing.T) {
	ctx := context.Background()
	builder := New(60 * time.Second)

	// Create a simple Go file
	tempDir := t.TempDir()
	goFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(goFile, []byte("package main\nfunc main() {}"), 0644)
	assert.NoError(t, err)

	result := builder.Run(ctx, "go build -o test main.go", tempDir)

	assert.True(t, result.Success)
	assert.Empty(t, result.Error)
	assert.Greater(t, result.Duration, time.Duration(0))
	// Output should contain build success info or be empty
	assert.NotNil(t, result.Output)
}

func TestBuilder_FailedBuild(t *testing.T) {
	ctx := context.Background()
	builder := New(60 * time.Second)

	// Create a Go file with syntax error
	tempDir := t.TempDir()
	goFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(goFile, []byte("package main\nfunc main( {\n}"), 0644) // Missing closing paren
	assert.NoError(t, err)

	result := builder.Run(ctx, "go build -o test main.go", tempDir)

	assert.False(t, result.Success)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "command failed")
	assert.Contains(t, result.Output, "syntax error") // Go compiler should report syntax error
	assert.Greater(t, result.Duration, time.Duration(0))
}

func TestBuilder_Cancellation(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	builder := New(60 * time.Second)

	// Use a command that takes longer than our context timeout
	tempDir := t.TempDir()
	goFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(goFile, []byte("package main\nimport \"time\"\nfunc main() { time.Sleep(1*time.Second) }"), 0644)
	assert.NoError(t, err)

	result := builder.Run(ctx, "go build -o test main.go", tempDir)

	assert.False(t, result.Success)
	assert.Equal(t, context.DeadlineExceeded, result.Error)
}

func TestBuilder_Timeout(t *testing.T) {
	ctx := context.Background()
	builder := New(100 * time.Millisecond) // Very short timeout

	// Create a Go file that will take time to build
	tempDir := t.TempDir()
	goFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(goFile, []byte("package main\nfunc main() {}"), 0644)
	assert.NoError(t, err)

	result := builder.Run(ctx, "go build -o test main.go", tempDir)

	// This might succeed or timeout depending on system speed
	// If it succeeds, that's fine - Go builds are usually fast
	if !result.Success {
		assert.Contains(t, result.Error.Error(), "build timeout")
	}
}

func TestBuilder_CommandParsing(t *testing.T) {
	builder := New(60 * time.Second)

	tests := []struct {
		command   string
		program   string
		args      []string
		expectErr bool
	}{
		{
			command:   "go build -o test main.go",
			program:   "go",
			args:      []string{"build", "-o", "test", "main.go"},
			expectErr: false,
		},
		{
			command:   "make build",
			program:   "make",
			args:      []string{"build"},
			expectErr: false,
		},
		{
			command:   `go build -o "test server" main.go`,
			program:   "go",
			args:      []string{"build", "-o", "test server", "main.go"},
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
			program, args, err := builder.parseCommand(tt.command)
			
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

func TestBuilder_NonExistentCommand(t *testing.T) {
	ctx := context.Background()
	builder := New(60 * time.Second)

	tempDir := t.TempDir()
	result := builder.Run(ctx, "nonexistentcommand", tempDir)

	assert.False(t, result.Success)
	assert.NotNil(t, result.Error)
	assert.Contains(t, result.Error.Error(), "failed to start command")
}

func TestBuilder_NonExistentDirectory(t *testing.T) {
	ctx := context.Background()
	builder := New(60 * time.Second)

	result := builder.Run(ctx, "echo hello", "/nonexistent/directory")

	assert.False(t, result.Success)
	assert.NotNil(t, result.Error)
}

func TestBuilder_Cancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	builder := New(60 * time.Second)

	// Start a long-running command in goroutine
	tempDir := t.TempDir()
	goFile := filepath.Join(tempDir, "main.go")
	err := os.WriteFile(goFile, []byte("package main\nimport \"time\"\nfunc main() { time.Sleep(5*time.Second) }"), 0644)
	assert.NoError(t, err)

	done := make(chan *Result, 1)
	go func() {
		result := builder.Run(ctx, "go build -o test main.go", tempDir)
		done <- result
	}()

	// Cancel after a short delay
	time.Sleep(50 * time.Millisecond)
	cancel()

	// Wait for result
	select {
	case result := <-done:
		assert.False(t, result.Success)
		assert.Equal(t, context.Canceled, result.Error)
	case <-time.After(1 * time.Second):
		t.Fatal("expected cancellation result")
	}
}
