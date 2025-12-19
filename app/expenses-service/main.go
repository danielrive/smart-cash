package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"smart-cash/expenses-service/internal/common"
	"smart-cash/expenses-service/internal/handler"
	"smart-cash/expenses-service/internal/handler/dto"
	"smart-cash/expenses-service/internal/repositories"
	"smart-cash/expenses-service/internal/service"
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
	expensesTable     string
	awsRegion         string
	jwtSecret         []byte
	notToLogEndpoints = []string{"/expenses/health", "/expenses/metrics"}
	logger            *slog.Logger
)

func init() {
	// Logger config - Initialize logger first
	logsConfig := logging.LoadConfig()
	logger = logging.InitLogger(logsConfig, "expenses-service")

	// Set ServiceName first for logger initialization
	common.ServiceName = os.Getenv("SERVICE_NAME")
	if common.ServiceName == "" {
		common.ServiceName = "expenses-service" // fallback for logger
	}

	common.DomainName = os.Getenv("DOMAIN_NAME")
	if common.DomainName == "" {
		common.DomainName = "localhost"
	}

	expensesTable = os.Getenv("DYNAMODB_EXPENSES_TABLE")
	if expensesTable == "" {
		logger.Error("environment variable not found", slog.String("variable", "DYNAMODB_EXPENSES_TABLE"))
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
		logger.Error("unable to load SDK config", slog.String("error", err.Error()))
	}
	// define uuid helper
	uuidHelper := utils.NewUUIDHelper()

	dynamoClient := dynamodb.NewFromConfig(cfg)
	// create a router with gin
	router := gin.New()

	router.Use(
		otelgin.Middleware(common.ServiceName, otelgin.WithFilter(filterTraces)),
		logging.HTTPMiddleware(logger, notToLogEndpoints),
		gin.Recovery(),
	)

	// // Initialize expenses repository
	expensesRepo := repositories.NewDynamoDBExpensesRepository(dynamoClient, expensesTable, logger)

	// Initialize expenses service
	expensesService := service.NewExpensesService(expensesRepo, uuidHelper, logger)

	// Init expenses handler
	expensesHandler := handler.NewExpensesHandler(expensesService, logger)

	// Public routes
	router.GET("/expenses/health", expensesHandler.HealthCheck)

	// Protected routes - All expense operations require authentication
	router.POST("/expenses",
		middleware.AuthMiddleware(jwtSecret),
		middleware.ValidateBody[dto.CreateExpenseRequest](), // Validate request body
		expensesHandler.CreateExpense)

	router.GET("/expenses/:expenseId",
		middleware.AuthMiddleware(jwtSecret),
		middleware.ValidatePathParam("expenseId", "uuid"), // Validate expenseId is a valid UUID
		expensesHandler.GetExpensesById)

	router.GET("/expenses",
		middleware.AuthMiddleware(jwtSecret),
		middleware.RequireOneOfQueryParams([]string{"userId", "category"}), // Require at least one query param
		expensesHandler.GetExpensesByQuery)

	router.DELETE("/expenses/:expenseId",
		middleware.AuthMiddleware(jwtSecret),
		middleware.ValidatePathParam("expenseId", "uuid"), // Validate expenseId is a valid UUID
		expensesHandler.DeleteExpense)

	router.Run(":8282")

}

func filterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}
