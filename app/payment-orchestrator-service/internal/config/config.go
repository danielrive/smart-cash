package config

import (
	"fmt"
	"log/slog"
	"os"
	"smart-cash/utils/logging"
)

type Config struct {
	ServiceName        string
	DomainName         string
	SQSQueueURL        string
	PaymentTable       string
	ExpensesServiceURL string
	BankServiceURL     string
	PaymentServiceURL  string
	OtelCollector      string
	AwsRegion          string
	JWTSecret          []byte
	Logger             *slog.Logger
}

func LoadConfig() (*Config, error) {
	serviceName := getEnv("SERVICE_NAME", "payment-orchestrator-service")

	// Initialize logger first so we can use it
	logCfg := logging.LoadConfig()
	logger := logging.InitLogger(logCfg, serviceName)

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	cfg := &Config{
		ServiceName:        serviceName,
		DomainName:         getEnv("DOMAIN_NAME", "localhost"),
		SQSQueueURL:        os.Getenv("SQS_QUEUE_URL"),
		PaymentTable:       os.Getenv("DYNAMODB_PAYMENT_TABLE"),
		ExpensesServiceURL: getEnv("EXPENSES_SERVICE_URL", "http://expenses.develop.svc.cluster.local:80"),
		BankServiceURL:     getEnv("BANK_SERVICE_URL", "http://bank.develop.svc.cluster.local:80"),
		PaymentServiceURL:  getEnv("PAYMENT_SERVICE_URL", "http://payment.develop.svc.cluster.local:80"),
		OtelCollector:      os.Getenv("OTEL_COLLECTOR"),
		AwsRegion:          os.Getenv("AWS_REGION"),
		JWTSecret:          []byte(jwtSecret),
		Logger:             logger,
	}

	// Validation logic
	if cfg.SQSQueueURL == "" {
		return nil, fmt.Errorf("SQS_QUEUE_URL is required")
	}
	if cfg.PaymentTable == "" {
		return nil, fmt.Errorf("DYNAMODB_PAYMENT_TABLE is required")
	}
	if cfg.OtelCollector == "" {
		return nil, fmt.Errorf("OTEL_COLLECTOR is required")
	}
	if cfg.AwsRegion == "" {
		return nil, fmt.Errorf("AWS_REGION is required")
	}

	return cfg, nil
}

// Helper for defaults
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
