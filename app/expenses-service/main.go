package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"smart-cash/expenses-service/internal/common"
	"smart-cash/expenses-service/internal/handler"
	"smart-cash/expenses-service/internal/repositories"
	"smart-cash/expenses-service/internal/service"
	"smart-cash/utils"

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
	notToLogEndpoints = []string{"/expenses/health", "/expenses/metrics"}
	logger            *slog.Logger
)

func init() {
	// validate ENV variables

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

	common.ServiceName = os.Getenv("SERVICE_NAME")
	if otelCollector == "" {
		logger.Error("environment variable not found", slog.String("variable", "SERVICE_NAME"))
		os.Exit(1)
	}

}

func main() {
	// Set-up logger handler
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // (Info, Warn, Error)
	}))
	slog.SetDefault(logger)

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
		gin.LoggerWithWriter(gin.DefaultWriter, "/expenses/health"),
		gin.Recovery(), gin.Recovery(),
	)

	// // Initialize expenses repository
	expensesRepo := repositories.NewDynamoDBExpensesRepository(dynamoClient, expensesTable, logger)

	// Initialize expenses service
	expensesService := service.NewExpensesService(expensesRepo, uuidHelper, logger)

	// Init expenses handler
	expensesHandler := handler.NewExpensesHandler(expensesService, logger)

	// create expenses
	router.POST("/expenses/", expensesHandler.CreateExpense)

	// define router for get expenses by tag
	router.GET("/expenses/:expenseId", expensesHandler.GetExpensesById)
	// define router for get expenses by category or userId
	router.GET("/expenses", expensesHandler.GetExpensesByQuery)

	router.DELETE("/expenses/:expenseId", expensesHandler.DeleteExpense)

	// Endpoint to test health check
	router.GET("/expenses/health", expensesHandler.HealthCheck)

	router.Run(":8282")

}

func filterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}
