package builder

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Result struct {
	Success   bool
	Output    string // Combined stdout + stderr
	Duration  time.Duration
	Error     error
}

type Builder struct {
	timeout time.Duration
	mu      sync.Mutex
	cmd     *exec.Cmd
}

// New creates a Builder with specified timeout
// Default timeout: 60 seconds
func New(timeout time.Duration) *Builder {
	return &Builder{
		timeout: timeout,
	}
}

// Run executes a build command with timeout and cancellation
// Captures combined stdout + stderr
// Returns Result with success status, output, and duration
func (b *Builder) Run(ctx context.Context, buildCmd string, workingDir string) *Result {
	start := time.Now()
	
	// Parse command string into program and args
	program, args, err := b.parseCommand(buildCmd)
	if err != nil {
		return &Result{
			Success:  false,
			Output:   "",
			Duration: time.Since(start),
			Error:    fmt.Errorf("failed to parse command: %w", err),
		}
	}

	// Create command with context
	cmd := exec.CommandContext(ctx, program, args...)
	cmd.Dir = workingDir

	// Set up output capture
	output, err := b.captureOutput(ctx, cmd)
	if err != nil {
		return &Result{
			Success:  false,
			Output:   output,
			Duration: time.Since(start),
			Error:    err,
		}
	}

	return &Result{
		Success:  true,
		Output:   output,
		Duration: time.Since(start),
		Error:    nil,
	}
}

// Cancel stops any in-progress build
func (b *Builder) Cancel() {
	b.mu.Lock()
	defer b.mu.Unlock()
	
	if b.cmd != nil && b.cmd.Process != nil {
		b.cmd.Process.Kill()
		b.cmd = nil
	}
}

// parseCommand splits command string into program and arguments
// Supports quoted arguments and shell-like escaping
func (b *Builder) parseCommand(command string) (string, []string, error) {
	// Simple parsing - split by spaces, respecting quotes
	var args []string
	var current strings.Builder
	inQuotes := false
	quoteChar := rune(0)

	for i, r := range command {
		switch {
		case r == '"' || r == '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = r
			} else if r == quoteChar {
				inQuotes = false
				quoteChar = 0
			} else {
				current.WriteRune(r)
			}
		case r == ' ' && !inQuotes:
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		default:
			current.WriteRune(r)
		}
		
		// Handle end of string
		if i == len(command)-1 && current.Len() > 0 {
			args = append(args, current.String())
		}
	}

	if len(args) == 0 {
		return "", nil, fmt.Errorf("empty command")
	}

	return args[0], args[1:], nil
}

// captureOutput runs command and captures combined stdout + stderr
func (b *Builder) captureOutput(ctx context.Context, cmd *exec.Cmd) (string, error) {
	// Create pipes for stdout and stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("failed to create stdout pipe: %w", err)
	}
	
	stderr, err := cmd.StderrPipe()
	if err != nil {
		stdout.Close()
		return "", fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Store command reference for cancellation
	b.mu.Lock()
	b.cmd = cmd
	b.mu.Unlock()

	// Start command
	if err := cmd.Start(); err != nil {
		stdout.Close()
		stderr.Close()
		b.mu.Lock()
		b.cmd = nil
		b.mu.Unlock()
		return "", fmt.Errorf("failed to start command: %w", err)
	}

	// Capture output with timeout
	outputChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	go func() {
		output, err := b.readOutput(stdout, stderr)
		select {
		case outputChan <- output:
		case <-ctx.Done():
		}
		select {
		case errorChan <- err:
		case <-ctx.Done():
		}
	}()

	// Wait for completion with timeout
	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	select {
	case err := <-done:
		// Command completed
		b.mu.Lock()
		b.cmd = nil
		b.mu.Unlock()
		
		output := <-outputChan
		cmdErr := <-errorChan
		
		if err != nil {
			return output, fmt.Errorf("command failed: %w", err)
		}
		if cmdErr != nil {
			return output, fmt.Errorf("failed to read output: %w", cmdErr)
		}
		return output, nil

	case <-ctx.Done():
		// Context cancelled
		b.Cancel()
		return "", ctx.Err()

	case <-time.After(b.timeout):
		// Timeout reached
		b.Cancel()
		return "", fmt.Errorf("build timeout after %v", b.timeout)
	}
}

// readOutput combines stdout and stderr into single string
func (b *Builder) readOutput(stdout, stderr io.Reader) (string, error) {
	var output strings.Builder
	
	// Create scanner for combined output
	combined := io.MultiReader(stdout, stderr)
	scanner := bufio.NewScanner(combined)
	
	for scanner.Scan() {
		output.WriteString(scanner.Text())
		output.WriteString("\n")
	}
	
	if err := scanner.Err(); err != nil {
		return output.String(), fmt.Errorf("failed to read output: %w", err)
	}
	
	return output.String(), nil
}
