package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"hotreload/internal/builder"
	"hotreload/internal/cli"
	"hotreload/internal/debouncer"
	"hotreload/internal/logger"
	"hotreload/internal/process"
	"hotreload/internal/watcher"
)

func main() {
	// Parse CLI arguments
	config, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		slog.Error("failed to parse arguments", "error", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		slog.Error("configuration validation failed", "error", err)
		os.Exit(1)
	}

	// Setup logging
	logLevel := slog.LevelInfo
	// Check for --verbose flag
	for i, arg := range os.Args {
		if arg == "--verbose" {
			logLevel = slog.LevelDebug
			break
		}
		// Remove --verbose from args for cleaner processing
		if arg == "--verbose" && i < len(os.Args)-1 {
			os.Args = append(os.Args[:i], os.Args[i+1:]...)
			break
		}
	}
	logger.Setup(logLevel)

	slog.Info("hotreload starting", "root", config.Root)
	slog.Info("config validated", 
		"build_cmd", config.BuildCmd, 
		"exec_cmd", config.ExecCmd)

	// Setup context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		slog.Info("received signal", "signal", sig)
		cancel()
	}()

	// Initialize watcher
	w, err := watcher.New(ctx, config.Root)
	if err != nil {
		slog.Error("failed to initialize watcher", "error", err)
		os.Exit(1)
	}

	// Initialize debouncer
	d := debouncer.New(300 * time.Millisecond)
	d.Start(ctx)

	// Initialize builder
	b := builder.New(60 * time.Second)

	// Initialize process manager
	p := process.New()

	// Start watching
	go w.Start(ctx)

	// Process events
	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down gracefully")
			return

		case event := <-w.Events():
			if watcher.IsGoFile(event.Path) {
				slog.Debug("go file changed", "path", event.Path, "op", event.Op)
				d.Trigger()
			}

		case <-d.Output():
			slog.Info("build starting", "cmd", config.BuildCmd)
			result := b.Run(ctx, config.BuildCmd, config.Root)
			
			if result.Success {
				slog.Info("build succeeded", 
					"duration", result.Duration.String(),
					"output", result.Output)
				
				// Start/restart server
				err := p.Start(ctx, config.ExecCmd, config.Root)
				if err != nil {
					slog.Error("failed to start server", "error", err)
				} else {
					slog.Info("server started", 
						"pid", p.PID(),
						"uptime", p.Uptime().String())
				}
			} else {
				slog.Error("build failed", 
					"duration", result.Duration.String(),
					"error", result.Error,
					"output", result.Output)
			}

		case err := <-w.Errors():
			slog.Error("watcher error", "error", err)
		}
	}
}
