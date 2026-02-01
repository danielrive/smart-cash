package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"smart-cash/bank-service/internal/common"
	appconfig "smart-cash/bank-service/internal/config"
	"smart-cash/bank-service/internal/handler"
	"smart-cash/bank-service/internal/handler/dto"
	"smart-cash/bank-service/internal/repositories"
	"smart-cash/bank-service/internal/service"
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
	notToLogEndpoints = []string{"/bank/health", "/bank/metrics"}
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

	router := gin.New()

	router.Use(
		otelgin.Middleware(cfg.ServiceName, otelgin.WithFilter(filterTraces)),
		logging.HTTPMiddleware(cfg.Logger, notToLogEndpoints),
		gin.Recovery(),
	)

	bankRepo := repositories.NewDynamoDBBankRepository(dynamoClient, cfg.BankTable, cfg.Logger)
	bankService := service.NewBankService(bankRepo, cfg.Logger)
	bankHandler := handler.NewBankHandler(bankService, cfg.Logger)

	router.GET("/bank/health", bankHandler.HealthCheck)

	router.POST("/bank/pay",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.ValidateBody[dto.PayExpenseRequest](),
		bankHandler.HandlePayment)

	router.GET("/bank/user/:userId",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.ValidatePathParam("userId", "uuid"),
		bankHandler.GetUser)

	return router.Run(":8585")
}

func filterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}
