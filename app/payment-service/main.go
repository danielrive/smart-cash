package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"smart-cash/payment-service/internal/common"
	appconfig "smart-cash/payment-service/internal/config"
	"smart-cash/payment-service/internal/handler"
	"smart-cash/payment-service/internal/handler/dto"
	"smart-cash/payment-service/internal/repositories"
	"smart-cash/payment-service/internal/service"
	"smart-cash/utils"
	"smart-cash/utils/logging"
	"smart-cash/utils/middleware"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
)

var (
	notToLogEndpoints = []string{"/payment/health", "/payment/metrics"}
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

	// Init OTel TracerProvider
	tp := utils.InitOpenTelemetry(cfg.OtelCollector, cfg.ServiceName, cfg.Logger)
	otel.SetTracerProvider(tp)

	// Configure AWS SDK
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.AwsRegion),
	)
	if err != nil {
		return fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	// Define AWS clients
	dynamoClient := dynamodb.NewFromConfig(awsCfg)

	router := gin.New()

	router.Use(
		otelgin.Middleware(cfg.ServiceName, otelgin.WithFilter(filterTraces)),
		logging.HTTPMiddleware(cfg.Logger, notToLogEndpoints),
		gin.Recovery(),
	)

	// UUID helper
	uuid := utils.NewUUIDHelper()

	// Initialize Payment repository
	paymentRepo := repositories.NewDynamoDBPaymentRepository(dynamoClient, cfg.PaymentTable, cfg.Logger)

	// Initialize Payment service
	paymentService := service.NewPaymentService(paymentRepo, uuid, cfg.Logger)

	// Init Payment handler
	paymentHandler := handler.NewPaymentHandler(paymentService, cfg.Logger)

	// Public routes
	router.GET("/payment/health", paymentHandler.HealthCheck)

	// Protected routes - All payment operations require authentication
	router.POST("/payment",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.ValidateBody[dto.ProcessPaymentRequest](), // Validate payment request
		paymentHandler.ProcessPayment)

	router.GET("/payment/:paymentId",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.ValidatePathParam("paymentId", "uuid"), // Validate paymentId is UUID
		paymentHandler.GetPayment)

	router.PUT("/payment/:paymentId/status",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.ValidatePathParam("paymentId", "uuid"),
		middleware.ValidateBody[dto.UpdatePaymentStatusRequest](),
		paymentHandler.UpdatePaymentStatus)

	return router.Run(":8989")
}

func filterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}
