package logging

import (
	"log/slog"
	"os"
	"strings"
)

// logging configs
type Config struct {
	Level slog.Level
}

// LoadConfig loads logging configuration from environment variables
func LoadConfig() *Config {
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		levelStr = "info"
	}

	level := parseLogLevel(strings.ToLower(levelStr))

	return &Config{
		Level: level,
	}
}

// parseLogLevel converts a string to slog.Level
func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
