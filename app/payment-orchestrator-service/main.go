package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
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
	"smart-cash/utils/logging"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
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
	expensesClient := clients.NewExpensesClient(cfg.ExpensesServiceURL, cfg.Logger, cfg.JWTSecret, cfg.ServiceName)
	bankClient := clients.NewBankClient(cfg.BankServiceURL, cfg.Logger, cfg.JWTSecret, cfg.ServiceName)
	paymentClient := clients.NewPaymentClient(cfg.PaymentServiceURL, cfg.Logger, cfg.JWTSecret, cfg.ServiceName)

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

	// Set up HTTP server for health checks (required by Kubernetes liveness/readiness probes)
	router := gin.New()
	router.Use(
		otelgin.Middleware(cfg.ServiceName),
		logging.HTTPMiddleware(cfg.Logger, []string{"/payment-orchestrator/health"}),
		gin.Recovery(),
	)

	// Health check endpoint
	router.GET("/payment-orchestrator/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	httpServer := &http.Server{
		Addr:    ":8088",
		Handler: router,
	}

	// Set up graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start HTTP server in a goroutine
	errChan := make(chan error, 2)
	go func() {
		cfg.Logger.Info("starting HTTP server for health checks", "port", "8088")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errChan <- fmt.Errorf("HTTP server error: %w", err)
		}
	}()

	// Start SQS consumer in a goroutine
	go func() {
		cfg.Logger.Info("starting SQS consumer")
		if err := sqsConsumer.Start(ctx); err != nil {
			// Context was cancelled (shutdown signal received) - this is expected
			if ctx.Err() != nil {
				cfg.Logger.Info("SQS consumer stopped gracefully",
					"reason", "context cancelled",
				)
				errChan <- nil
				return
			}
			// Some other unexpected error occurred
			cfg.Logger.Error("SQS consumer stopped unexpectedly",
				"error", err,
			)
			errChan <- fmt.Errorf("SQS consumer stopped: %w", err)
		} else {
			errChan <- nil
		}
	}()

	// Wait for shutdown signal or error
	select {
	case <-ctx.Done():
		cfg.Logger.Info("shutdown signal received, shutting down gracefully")
	case err := <-errChan:
		if err != nil {
			cfg.Logger.Error("service error, shutting down", "error", err)
			// Shutdown HTTP server
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := httpServer.Shutdown(shutdownCtx); err != nil {
				cfg.Logger.Error("error shutting down HTTP server", "error", err)
			}
			return err
		}
	}

	// Graceful shutdown: stop HTTP server
	cfg.Logger.Info("shutting down HTTP server")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		cfg.Logger.Error("error shutting down HTTP server", "error", err)
	}

	cfg.Logger.Info("payment orchestrator service stopped gracefully")
	return nil
}
