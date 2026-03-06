package process

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Status int

const (
	StatusStopped Status = iota
	StatusStarting
	StatusRunning
	StatusStopping
	StatusFailed
)

func (s Status) String() string {
	switch s {
	case StatusStopped:
		return "stopped"
	case StatusStarting:
		return "starting"
	case StatusRunning:
		return "running"
	case StatusStopping:
		return "stopping"
	case StatusFailed:
		return "failed"
	default:
		return "unknown"
	}
}

type Process struct {
	cmd      *exec.Cmd
	status   Status
	pid      int
	startTime time.Time
	mu       sync.RWMutex
	cancel   context.CancelFunc
}

type Manager struct {
	process *Process
	mu      sync.RWMutex
}

// New creates a new Process Manager
func New() *Manager {
	return &Manager{
		process: &Process{
			status: StatusStopped,
		},
	}
}

// Start executes the command and manages the process
func (m *Manager) Start(ctx context.Context, execCmd string, workingDir string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Parse command
	program, args, err := m.parseCommand(execCmd)
	if err != nil {
		return fmt.Errorf("failed to parse command: %w", err)
	}

	// Stop existing process if running
	if m.process.status == StatusRunning {
		if err := m.stopInternal(); err != nil {
			return fmt.Errorf("failed to stop existing process: %w", err)
		}
	}

	// Create context for this process
	processCtx, cancel := context.WithCancel(ctx)
	m.process.cancel = cancel

	// Setup command
	cmd := exec.CommandContext(processCtx, program, args...)
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Set process group for proper signal handling
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	// Update process state
	m.process.cmd = cmd
	m.process.status = StatusStarting

	// Start the process
	if err := cmd.Start(); err != nil {
		m.process.status = StatusFailed
		cancel()
		return fmt.Errorf("failed to start process: %w", err)
	}

	m.process.pid = cmd.Process.Pid
	m.process.startTime = time.Now()
	m.process.status = StatusRunning

	// Monitor process in goroutine
	go m.monitorProcess(processCtx)

	return nil
}

// Stop gracefully terminates the process
func (m *Manager) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return m.stopInternal()
}

// Status returns the current process status
func (m *Manager) Status() Status {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.process.status
}

// PID returns the current process PID
func (m *Manager) PID() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.process.pid
}

// Uptime returns how long the process has been running
func (m *Manager) Uptime() time.Duration {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	if m.process.status != StatusRunning {
		return 0
	}
	return time.Since(m.process.startTime)
}

// IsRunning returns true if process is currently running
func (m *Manager) IsRunning() bool {
	return m.Status() == StatusRunning
}

// stopInternal is the internal stop method (must be called with lock held)
func (m *Manager) stopInternal() error {
	if m.process.status != StatusRunning {
		return nil
	}

	m.process.status = StatusStopping

	// Try graceful shutdown first (SIGTERM)
	if m.process.cmd != nil && m.process.cmd.Process != nil {
		// Send SIGTERM to process group
		pgid, err := syscall.Getpgid(m.process.pid)
		if err == nil {
			syscall.Kill(-pgid, syscall.SIGTERM)
		} else {
			m.process.cmd.Process.Signal(syscall.SIGTERM)
		}

		// Wait for graceful shutdown
		done := make(chan error, 1)
		go func() {
			done <- m.process.cmd.Wait()
		}()

		select {
		case <-done:
			// Process terminated gracefully
			m.process.status = StatusStopped
			if m.process.cancel != nil {
				m.process.cancel()
			}
			return nil

		case <-time.After(5 * time.Second):
			// Force kill if graceful shutdown fails
			if pgid, err := syscall.Getpgid(m.process.pid); err == nil {
				syscall.Kill(-pgid, syscall.SIGKILL)
			} else {
				m.process.cmd.Process.Kill()
			}
			m.process.cmd.Wait()
			m.process.status = StatusStopped
			if m.process.cancel != nil {
				m.process.cancel()
			}
			return fmt.Errorf("process did not terminate gracefully, force killed")
		}
	}

	return nil
}

// monitorProcess monitors the running process and updates status
func (m *Manager) monitorProcess(ctx context.Context) {
	if m.process.cmd == nil {
		return
	}

	// Wait for process to exit
	_ = m.process.cmd.Wait()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Update status based on exit reason
	select {
	case <-ctx.Done():
		// Context cancelled, expected shutdown
		m.process.status = StatusStopped
	default:
		// Process exited unexpectedly
		if m.process.status == StatusRunning {
			m.process.status = StatusFailed
		}
	}

	// Clean up
	if m.process.cancel != nil {
		m.process.cancel()
		m.process.cancel = nil
	}
}

// parseCommand splits command string into program and arguments
func (m *Manager) parseCommand(command string) (string, []string, error) {
	var args []string
	var current strings.Builder
	inQuotes := bool(false)
	quoteChar := rune(0)

	for i, r := range command {
		switch {
		case (r == '"' || r == '\'') && !inQuotes:
			inQuotes = true
			quoteChar = r
		case r == quoteChar && inQuotes:
			inQuotes = false
			quoteChar = 0
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
