package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"smart-cash/bank-service/internal/handler"
	"smart-cash/bank-service/internal/repositories"
	"smart-cash/bank-service/internal/service"
	"smart-cash/utils"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
)

var logger *slog.Logger

var notToLogEndpoints = []string{"/bank/health", "/bank/metrics"}

func main() {
	// Set-up logger handler
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // (Info, Warn, Error)
	}))
	slog.SetDefault(logger)

	// Init OTel TracerProvider
	tp := utils.InitOpenTelemetry(os.Getenv("OTEL_COLLECTOR"), os.Getenv("SERVICE_NAME"), logger)

	otel.SetTracerProvider(tp)

	// configure the SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-west-2"),
	)

	bankTable := os.Getenv("DYNAMODB_BANK_TABLE")
	if bankTable == "" {
		logger.Error("environment variable not found", slog.String("variable", "DYNAMODB_BANK_TABLE"))
		os.Exit(1)
	}

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
		otelgin.Middleware(os.Getenv("OTEL_COLLECTOR"), otelgin.WithFilter(filterTraces)),
		gin.LoggerWithWriter(gin.DefaultWriter, "/bank/health"),
		gin.Recovery(), gin.Recovery(),
	)
	// // Initialize bank repository
	bankRepo := repositories.NewDynamoDBBankRepository(dynamoClient, bankTable, logger) // Harcoded dynamotable to use data already uploaded

	// Initialize bank service
	bankService := service.NewBankService(bankRepo, logger)

	// Init bank handler
	bankHandler := handler.NewBankHandler(bankService, logger)

	// create bank
	router.POST("/bank/pay", bankHandler.HandlePayment)

	// Endpoint to test health check
	router.GET("/bank/health", bankHandler.HealthCheck)

	router.Run(":8585")

}

func filterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}
