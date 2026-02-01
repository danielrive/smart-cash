package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"smart-cash/payment-orchestrator-service/internal/clients"
	"smart-cash/payment-orchestrator-service/internal/common"
	appconfig "smart-cash/payment-orchestrator-service/internal/config"
	"smart-cash/payment-orchestrator-service/internal/consumers"
	"smart-cash/payment-orchestrator-service/internal/service"
	"smart-cash/utils"

	"go.opentelemetry.io/otel"
)

func main() {
	if err := run(); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

// run contains the main application logic and returns an error instead of exiting.
// This makes it testable and allows for better error handling.
func run() error {
	// Load application configuration
	cfg, err := appconfig.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Set common variables from config
	common.ServiceName = cfg.ServiceName
	common.DomainName = cfg.DomainName

	cfg.Logger.Info("payment orchestrator service starting",
		"service_name", cfg.ServiceName,
		"sqs_queue_url", cfg.SQSQueueURL,
		"expenses_service_url", cfg.ExpensesServiceURL,
		"bank_service_url", cfg.BankServiceURL,
		"payment_service_url", cfg.PaymentServiceURL,
	)

	// Initialize OpenTelemetry
	tp := utils.InitOpenTelemetry(cfg.OtelCollector, cfg.ServiceName, cfg.Logger)
	otel.SetTracerProvider(tp)
	defer func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := utils.ShutdownTracerProvider(ctx, tp, cfg.Logger); err != nil {
			cfg.Logger.Error("failed to shutdown tracer provider", "error", err)
		}
	}()

	// Initialize HTTP clients
	expensesClient := clients.NewExpensesClient(cfg.ExpensesServiceURL, cfg.Logger)
	bankClient := clients.NewBankClient(cfg.BankServiceURL, cfg.Logger)
	paymentClient := clients.NewPaymentClient(cfg.PaymentServiceURL, cfg.Logger)

	// Initialize orchestrator service
	orchestratorService := service.NewOrchestratorService(
		expensesClient,
		bankClient,
		paymentClient,
		cfg.Logger,
	)

	// Initialize SQS consumer
	sqsConsumer, err := consumers.NewSQSConsumer(
		cfg.SQSQueueURL,
		cfg.AwsRegion,
		cfg.Logger,
		orchestratorService,
	)
	if err != nil {
		return fmt.Errorf("failed to create SQS consumer: %w", err)
	}

	// Set up graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start consuming messages - this blocks until context is cancelled (shutdown signal)
	// The Start() method checks ctx.Err() and returns when context is cancelled
	cfg.Logger.Info("starting SQS consumer")
	if err := sqsConsumer.Start(ctx); err != nil {
		// Context was cancelled (shutdown signal received)
		if ctx.Err() != nil {
			cfg.Logger.Info("payment orchestrator service stopped")
			return nil
		}
		// Some other error occurred
		cfg.Logger.Error("SQS consumer error", "error", err)
		return fmt.Errorf("SQS consumer error: %w", err)
	}

	// This should never be reached, but included for completeness
	return nil
}
