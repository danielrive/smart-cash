package main

import (
	"context"
	"log/slog"
	"os"
	"smart-cash/bank-service/internal/handler"
	"smart-cash/bank-service/internal/repositories"
	"smart-cash/bank-service/internal/service"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
)

func main() {
	// Set-up logger handler
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // (Info, Warn, Error)
	}))
	slog.SetDefault(logger)

	// configure the SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("us-west-2"),
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
		gin.LoggerWithWriter(gin.DefaultWriter, "/bank/health"),
		gin.Recovery(),
	)
	// // Initialize bank repository
	bankRepo := repositories.NewDynamoDBBankRepository(dynamoClient, "bank-test", logger) // Harcoded dynamotable to use data already uploaded

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
