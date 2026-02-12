package config

import (
	"fmt"
	"log/slog"
	"os"
	"smart-cash/utils/logging"
	"smart-cash/utils/secrets"
)

type Config struct {
	ServiceName   string
	DomainName    string
	UsersTable    string
	OtelCollector string
	AwsRegion     string
	JWTSecret     []byte
	Logger        *slog.Logger
}

func LoadConfig() (*Config, error) {
	serviceName := getEnv("SERVICE_NAME", "user-service")

	// Initialize logger first so we can use it
	logCfg := logging.LoadConfig()
	logger := logging.InitLogger(logCfg, serviceName)

	cfg := &Config{
		ServiceName:   serviceName,
		DomainName:    getEnv("DOMAIN_NAME", "localhost"),
		UsersTable:    os.Getenv("DYNAMODB_USER_TABLE"),
		OtelCollector: os.Getenv("OTEL_COLLECTOR"),
		AwsRegion:     os.Getenv("AWS_REGION"),
		Logger:        logger,
	}

	// Validation logic
	if cfg.UsersTable == "" {
		return nil, fmt.Errorf("DYNAMODB_USER_TABLE variable is required")
	}
	if cfg.OtelCollector == "" {
		return nil, fmt.Errorf("OTEL_COLLECTOR variable is required")
	}
	if cfg.AwsRegion == "" {
		return nil, fmt.Errorf("AWS_REGION variable is required")
	}

	// Read JWT secret from mounted file (Secrets Store CSI Driver)
	secret, err := secrets.ReadJWTSecret()
	if err != nil {
		return nil, fmt.Errorf("failed to read JWT secret: %w", err)
	}
	cfg.JWTSecret = secret

	return cfg, nil
}

// Helper for defaults
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
