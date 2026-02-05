package config

import (
	"fmt"
	"log/slog"
	"os"
	"smart-cash/utils/logging"
)

type Config struct {
	ServiceName   string
	DomainName    string
	PaymentTable  string
	OtelCollector string
	AwsRegion     string
	JWTSecret     []byte
	Logger        *slog.Logger
}

func LoadConfig() (*Config, error) {
	serviceName := getEnv("SERVICE_NAME", "payment-service")

	// Initialize logger first so we can use it
	logCfg := logging.LoadConfig()
	logger := logging.InitLogger(logCfg, serviceName)

	cfg := &Config{
		ServiceName:   serviceName,
		DomainName:    getEnv("DOMAIN_NAME", "localhost"),
		PaymentTable:  os.Getenv("DYNAMODB_PAYMENT_TABLE"),
		OtelCollector: os.Getenv("OTEL_COLLECTOR"),
		AwsRegion:     os.Getenv("AWS_REGION"),
		Logger:        logger,
	}

	// Validation logic
	if cfg.PaymentTable == "" {
		return nil, fmt.Errorf("DYNAMODB_PAYMENT_TABLE is required")
	}
	if cfg.OtelCollector == "" {
		return nil, fmt.Errorf("OTEL_COLLECTOR is required")
	}
	if cfg.AwsRegion == "" {
		return nil, fmt.Errorf("AWS_REGION is required")
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}
	cfg.JWTSecret = []byte(secret)

	return cfg, nil
}

// Helper for defaults
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
