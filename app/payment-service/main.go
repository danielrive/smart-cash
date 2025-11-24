package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"smart-cash/payment-service/internal/common"
	"smart-cash/payment-service/internal/handler"
	"smart-cash/payment-service/internal/handler/dto"
	"smart-cash/payment-service/internal/repositories"
	"smart-cash/payment-service/internal/service"
	"smart-cash/utils"
	"smart-cash/utils/logging"
	"smart-cash/utils/middleware"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
)

var (
	otelCollector     string
	paymentTable      string
	awsRegion         string
	jwtSecret         []byte
	notToLogEndpoints = []string{"/payment/health", "/payment/metrics"}
	logger            *slog.Logger
	domainName        string
	sqsQueueURL       string
)

func init() {
	// Set ServiceName first for logger initialization
	common.ServiceName = os.Getenv("SERVICE_NAME")
	if common.ServiceName == "" {
		common.ServiceName = "payment-service"

		logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug, // (Info, Warn, Error)
		}))
		slog.SetDefault(logger)

		// validate ENV variables
		common.DomainName = os.Getenv("DOMAIN_NAME")
		if common.DomainName == "" {
			common.DomainName = "localhost"
		}

		paymentTable = os.Getenv("DYNAMODB_PAYMENT_TABLE")
		if paymentTable == "" {
			logger.Error("environment variable not found", slog.String("variable", "DYNAMODB_PAYMENT_TABLE"))
			os.Exit(1)
		}

		otelCollector = os.Getenv("OTEL_COLLECTOR")
		if otelCollector == "" {
			logger.Error("environment variable not found", slog.String("variable", "OTEL_COLLECTOR"))
			os.Exit(1)
		}

		awsRegion = os.Getenv("AWS_REGION")
		if awsRegion == "" {
			logger.Error("environment variable not found", slog.String("variable", "AWS_REGION"))
			os.Exit(1)
		}

		// Re-validate ServiceName (in case it was set via env)
		common.ServiceName = os.Getenv("SERVICE_NAME")
		if common.ServiceName == "" {
			logger.Error("environment variable not found", slog.String("variable", "SERVICE_NAME"))
			os.Exit(1)
		}

		jwtSecretStr := os.Getenv("JWT_SECRET")
		if jwtSecretStr == "" {
			logger.Warn("JWT_SECRET not set, using default (NOT FOR PRODUCTION!)")
			jwtSecretStr = "default-secret-change-me"
		}
		jwtSecret = []byte(jwtSecretStr)

	}
}

func main() {
	// Init OTel TracerProvider
	tp := utils.InitOpenTelemetry(otelCollector, common.ServiceName, logger)

	otel.SetTracerProvider(tp)

	// configure the SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
	)

	if err != nil {
		slog.Error("unable to load SDK config",
			"error", err.Error())
		os.Exit(1)
	}

	// Define aws clients
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// create a router with gin
	router := gin.New()

	router.Use(
		otelgin.Middleware(common.ServiceName, otelgin.WithFilter(filterTraces)),
		logging.HTTPMiddleware(logger, notToLogEndpoints),
		gin.Recovery(),
	)

	// uuid helper
	uuid := utils.NewUUIDHelper()

	// Initialize Payment repository
	paymentRepo := repositories.NewDynamoDBPaymentRepository(dynamoClient, paymentTable, logger) // Harcoded dynamotable to use data already uploaded

	// Initialize Payment service
	paymentService := service.NewPaymentService(paymentRepo, uuid, logger)

	// Init Payment handler
	paymentHandler := handler.NewPaymentHandler(paymentService, logger)

	// Public routes
	router.GET("/payment/health", paymentHandler.HealthCheck)

	// Protected routes - All payment operations require authentication
	router.POST("/payment",
		middleware.AuthMiddleware(jwtSecret),
		middleware.ValidateBody[dto.ProcessPaymentRequest](), // Validate payment request
		paymentHandler.ProcessPayment)

	router.GET("/payment/:transactionId",
		middleware.AuthMiddleware(jwtSecret),
		middleware.ValidatePathParam("transactionId", "uuid"), // Validate transactionId is UUID
		paymentHandler.GetPayment)

	router.Run(":8989")

}

func filterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}
