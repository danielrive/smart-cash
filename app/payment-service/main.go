package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"slices"
	"smart-cash/payment-service/internal/common"
	"smart-cash/payment-service/internal/handler"
	"smart-cash/payment-service/internal/repositories"
	"smart-cash/payment-service/internal/service"
	"smart-cash/utils"

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
	notToLogEndpoints = []string{"/payment/health", "/payment/metrics"}
	logger            *slog.Logger
	domainName        string
)

func init() {
	// start logger

	// Init OTel TracerProvider
	tp := utils.InitOpenTelemetry(otelCollector, common.ServiceName, logger)

	otel.SetTracerProvider(tp)

	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // (Info, Warn, Error)
	}))
	slog.SetDefault(logger)
	// validate ENV variables
	common.DomainName = os.Getenv("DOMAIN_NAME")
	if domainName == "" {
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

	common.ServiceName = os.Getenv("SERVICE_NAME")
	if otelCollector == "" {
		logger.Error("environment variable not found", slog.String("variable", "SERVICE_NAME"))
		os.Exit(1)
	}

}

func main() {

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
		gin.LoggerWithWriter(gin.DefaultWriter, "/payment/health"),
		gin.Recovery(), gin.Recovery(),
	)

	// uuid helper
	uuid := utils.NewUUIDHelper()

	// Initialize Payment repository
	paymentRepo := repositories.NewDynamoDBPaymentRepository(dynamoClient, paymentTable, logger) // Harcoded dynamotable to use data already uploaded

	// Initialize Payment service
	paymentService := service.NewPaymentService(paymentRepo, uuid, logger)

	// Init Payment handler
	paymentHandler := handler.NewPaymentHandler(paymentService, logger)

	// create Payment
	router.GET("/payment/:transactionId", paymentHandler.GetTransaction)

	// create Payment
	router.POST("/payment", paymentHandler.ProcessPayment)

	// Endpoint to test health check
	router.GET("/payment/health", paymentHandler.HealthCheck)

	router.Run(":8989")

}

func filterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}
