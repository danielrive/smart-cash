package logging

import (
	"log/slog"
	"os"
)

// InitLogger initializes and returns a configured JSON logger
// Always uses JSON format for structured logging
func InitLogger(config *Config, serviceName string) *slog.Logger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level:     config.Level,
		AddSource: false, // Set to true if you want file:line in logs
	})

	logger := slog.New(handler).With(
		slog.String("service", serviceName),
	)

	slog.SetDefault(logger)

	return logger
}
