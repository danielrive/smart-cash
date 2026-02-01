package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"smart-cash/expenses-service/internal/common"
	appconfig "smart-cash/expenses-service/internal/config"
	"smart-cash/expenses-service/internal/handler"
	"smart-cash/expenses-service/internal/handler/dto"
	"smart-cash/expenses-service/internal/repositories"
	"smart-cash/expenses-service/internal/service"
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
	notToLogEndpoints = []string{"/expenses/health", "/expenses/metrics"}
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
	cfg, err := appconfig.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	common.ServiceName = cfg.ServiceName
	common.DomainName = cfg.DomainName

	tp := utils.InitOpenTelemetry(cfg.OtelCollector, cfg.ServiceName, cfg.Logger)
	otel.SetTracerProvider(tp)

	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.AwsRegion),
	)
	if err != nil {
		return fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	dynamoClient := dynamodb.NewFromConfig(awsCfg)
	uuidHelper := utils.NewUUIDHelper()

	router := gin.New()

	router.Use(
		otelgin.Middleware(cfg.ServiceName, otelgin.WithFilter(filterTraces)),
		logging.HTTPMiddleware(cfg.Logger, notToLogEndpoints),
		gin.Recovery(),
	)

	expensesRepo := repositories.NewDynamoDBExpensesRepository(dynamoClient, cfg.ExpensesTable, cfg.Logger)
	expensesService := service.NewExpensesService(expensesRepo, uuidHelper, cfg.Logger)
	expensesHandler := handler.NewExpensesHandler(expensesService, cfg.Logger)

	router.GET("/expenses/health", expensesHandler.HealthCheck)

	router.POST("/expenses",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.ValidateBody[dto.CreateExpenseRequest](),
		expensesHandler.CreateExpense)

	router.GET("/expenses/:expenseId",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.ValidatePathParam("expenseId", "uuid"),
		expensesHandler.GetExpensesById)

	router.GET("/expenses",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.RequireOneOfQueryParams([]string{"userId", "category"}),
		expensesHandler.GetExpensesByQuery)

	router.DELETE("/expenses/:expenseId",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.ValidatePathParam("expenseId", "uuid"),
		expensesHandler.DeleteExpense)

	router.PUT("/expenses/:expenseId/status",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.ValidatePathParam("expenseId", "uuid"),
		middleware.ValidateBody[dto.UpdateExpenseStatusRequest](),
		expensesHandler.UpdateExpenseStatus)

	return router.Run(":8282")
}

func filterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}
