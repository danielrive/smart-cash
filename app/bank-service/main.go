package main

import (
	"context"
	"log"
	"os"
	"smart-cash/bank-service/internal/handler"
	"smart-cash/bank-service/internal/repositories"
	"smart-cash/bank-service/internal/service"

	"smart-cash/utils"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
)

func main() {
	// validate if env variables exists
	transactionsTable := os.Getenv("DYNAMODB_TRANSACTIONS_TABLE")
	if transactionsTable == "" {
		panic("DYNAMODB_TRANSACTIONS_TABLE cannot be empty")
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		panic("AWS_REGION cannot be empty")
	}

	// configure the SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}
	uuidHelper := utils.NewUUIDHelper()

	dynamoClient := dynamodb.NewFromConfig(cfg)
	// create a router with gin
	router := gin.New()
	router.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/health"),
		gin.Recovery(),
	)

	// // Initialize bank repository
	transactionsRepo := repositories.NewDynamoDBTransactionRepository(dynamoClient, transactionsTable, uuidHelper)

	// Initialize bank service
	bankService := service.NewBankService(transactionsRepo)

	// Init bank handler
	bankHandler := handler.NewBankHandler(bankService)

	router.POST("/", bankHandler.CreateTransaction)

	router.GET("/", bankHandler.GetTransactions)

	router.Run(":9582")

}
