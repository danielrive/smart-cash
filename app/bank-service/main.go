package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"smart-cash/bank-service/internal/common"
	"smart-cash/bank-service/internal/handler"
	"smart-cash/bank-service/internal/handler/dto"
	"smart-cash/bank-service/internal/repositories"
	"smart-cash/bank-service/internal/service"
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
	logger            *slog.Logger
	domainName        string
	bankTable         string
	awsRegion         string
	jwtSecret         []byte
	notToLogEndpoints = []string{"/bank/health", "/bank/metrics"}
	otelCollector     string
)

func init() {
	// Set ServiceName first for logger initialization
	common.ServiceName = os.Getenv("SERVICE_NAME")
	if common.ServiceName == "" {
		common.ServiceName = "bank-service" // fallback for logger
	}

	// Logger config
	logsConfig := logging.LoadConfig()
	logger = logging.InitLogger(logsConfig, common.ServiceName)

	// validate ENV variables
	common.DomainName = os.Getenv("DOMAIN_NAME")
	if common.DomainName == "" {
		common.DomainName = "localhost"
	}

	bankTable = os.Getenv("DYNAMODB_BANK_TABLE")
	if bankTable == "" {
		logger.Error("environment variable not found", slog.String("variable", "DYNAMODB_BANK_TABLE"))
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
	// define uuid helper
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// create a router with gin
	router := gin.New()

	router.Use(
		otelgin.Middleware(common.ServiceName, otelgin.WithFilter(filterTraces)),
		logging.HTTPMiddleware(logger, notToLogEndpoints),
		gin.Recovery(),
	)
	// // Initialize bank repository
	bankRepo := repositories.NewDynamoDBBankRepository(dynamoClient, bankTable, logger) // Harcoded dynamotable to use data already uploaded

	// Initialize bank service
	bankService := service.NewBankService(bankRepo, logger)

	// Init bank handler
	bankHandler := handler.NewBankHandler(bankService, logger)

	// Public routes
	router.GET("/bank/health", bankHandler.HealthCheck)

	// Protected routes - All bank operations require authentication
	router.POST("/bank/pay",
		middleware.AuthMiddleware(jwtSecret),
		middleware.ValidateBody[dto.PayExpenseRequest](), // Validate payment request
		bankHandler.HandlePayment)

	router.GET("/bank/user/:userId",
		middleware.AuthMiddleware(jwtSecret),
		middleware.ValidatePathParam("userId", "uuid"), // Validate userId is UUID
		bankHandler.GetUser)
	router.Run(":8585")
}

func filterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}
