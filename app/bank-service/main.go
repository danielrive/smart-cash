package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"smart-cash/bank-service/internal/common"
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

var (
	logger            *slog.Logger
	domainName        string
	bankTable         string
	awsRegion         string
	notToLogEndpoints = []string{"/bank/health", "/bank/metrics"}
	otelCollector     string
)

func init() {
	// validate ENV variables
	common.DomainName = os.Getenv("DOMAIN_NAME")
	if domainName == "" {
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

	common.ServiceName = os.Getenv("SERVICE_NAME")

	if otelCollector == "" {
		logger.Error("environment variable not found", slog.String("variable", "SERVICE_NAME"))
		os.Exit(1)
	}

}

func main() {
	// Set-up logger handler
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
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
		slog.Error("unable to load SDK config",
			"error", err.Error())
		os.Exit(1)
	}
	// define uuid helper
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// create a router with gin
	router := gin.New()

	router.Use(
		otelgin.Middleware(otelCollector, otelgin.WithFilter(filterTraces)),
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

	// Get user saldo
	router.GET("/bank/user", bankHandler.GetUser)

	// Endpoint to test health check
	router.GET("/bank/health", bankHandler.HealthCheck)
	router.Run(":8585")
}

func filterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}
