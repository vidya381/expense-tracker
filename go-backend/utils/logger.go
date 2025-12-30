package utils

import (
	"log/slog"
	"os"
)

var Logger *slog.Logger

// InitLogger initializes the global logger with JSON formatting for production
// or text formatting for development based on environment.
func InitLogger() {
	env := os.Getenv("ENV")

	var handler slog.Handler
	if env == "production" {
		// JSON format for production (easier to parse by log aggregators)
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	} else {
		// Text format for development (more readable)
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	Logger = slog.New(handler)
	slog.SetDefault(Logger)
}

// LogInfo logs an info level message with optional key-value pairs
func LogInfo(msg string, args ...any) {
	slog.Info(msg, args...)
}

// LogDebug logs a debug level message with optional key-value pairs
func LogDebug(msg string, args ...any) {
	slog.Debug(msg, args...)
}

// LogError logs an error level message with optional key-value pairs
func LogError(msg string, args ...any) {
	slog.Error(msg, args...)
}

// LogWarn logs a warning level message with optional key-value pairs
func LogWarn(msg string, args ...any) {
	slog.Warn(msg, args...)
}
