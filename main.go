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
	"hotreload/internal/crashguard"
	"hotreload/internal/debouncer"
	"hotreload/internal/logger"
	"hotreload/internal/process"
	"hotreload/internal/watcher"
)

func main() {
	config, err := cli.ParseArgs(os.Args[1:])
	if err != nil {
		slog.Error("failed to parse arguments", "error", err)
		os.Exit(1)
	}

	if err := config.Validate(); err != nil {
		slog.Error("configuration validation failed", "error", err)
		os.Exit(1)
	}

	logLevel := slog.LevelInfo
	for i, arg := range os.Args {
		if arg == "--verbose" {
			logLevel = slog.LevelDebug
			break
		}
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

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		slog.Info("received signal", "signal", sig)
		cancel()
	}()

	w, err := watcher.New(ctx, config.Root)
	if err != nil {
		slog.Error("failed to initialize watcher", "error", err)
		os.Exit(1)
	}

	d := debouncer.New(300 * time.Millisecond)
	d.Start(ctx)

	b := builder.New(60 * time.Second)

	cg := crashguard.New(crashguard.Config{
		MaxRestarts: 5,
		BaseDelay:  1 * time.Second,
		MaxDelay:   30 * time.Second,
		Window:     5 * time.Minute,
	})

	p := process.New()

	go w.Start(ctx)

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
				
				if cg.ShouldRestart() {
					err := p.Start(ctx, config.ExecCmd, config.Root)
					if err != nil {
						slog.Error("failed to start server", "error", err)
						cg.RecordCrash()
						stats := cg.GetStats()
						slog.Warn("server crash recorded", 
							"crash_count", stats.CrashCount,
							"backoff_delay", stats.BackoffDelay.String(),
							"can_restart", stats.CanRestart)
						
						if stats.BackoffDelay > 0 {
							slog.Info("waiting for backoff delay", "delay", stats.BackoffDelay.String())
							select {
							case <-time.After(stats.BackoffDelay):
							case <-ctx.Done():
								return
							}
						}
					} else {
						slog.Info("server started", 
							"pid", p.PID(),
							"uptime", p.Uptime().String())
						cg.Reset()
					}
				} else {
					stats := cg.GetStats()
					slog.Error("restart blocked", 
						"crash_count", stats.CrashCount,
						"max_restarts", 5,
						"backoff_delay", stats.BackoffDelay.String())
					slog.Info("manual intervention required - server crashed too many times")
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
