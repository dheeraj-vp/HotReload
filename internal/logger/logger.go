package logger

import (
	"log/slog"
	"os"
)

var globalLogger *slog.Logger

// Setup initializes global slog logger with JSON format
// Level: DEBUG for development, INFO for production
func Setup(level slog.Level) {
	globalLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))
	slog.SetDefault(globalLogger)
}

// Init returns the global logger
// Must be called before any logging
func Init() *slog.Logger {
	if globalLogger == nil {
		// Default to INFO level if not initialized
		Setup(slog.LevelInfo)
	}
	return globalLogger
}
