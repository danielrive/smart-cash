package logging

import (
	"log/slog"
	"os"
)

// InitLogger initializes and returns a configured logger
func InitLogger(config *Config) *slog.Logger {
	var handler slog.Handler

	handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: config.Level,
	})

	logger := slog.New(handler)
	slog.SetDefault(logger)

	return logger
}
