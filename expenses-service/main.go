package main

import (
	"context"
	"log"
	"os"

	"expenses-service/internal/handler"

	"expenses-service/internal/repositories"
	"expenses-service/internal/service"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
)

func main() {
	// validate if env variables exists
	expensesTable := os.Getenv("DYNAMODB_EXPENSES_TABLE")
	if expensesTable == "" {
		panic("DYNAMODB_EXPENSES_TABLE cannot be empty")
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
	dynamoClient := dynamodb.NewFromConfig(cfg)
	// create a router with gin
	router := gin.Default()

	// // Initialize expenses repository
	expensesRepo := repositories.NewDynamoDBExpensesRepository(dynamoClient, expensesTable)
	// Initialize expenses service
	expensesService := service.NewExpensesService(expensesRepo)

	// Init expenses handler
	expensesHandler := handler.NewExpensesHandler(expensesService)

	router.POST("/", expensesHandler.CreateExpense)

	//router.GET("/calculateTotal", expensesHandler.CalculateTotalPerCategory)

	// define router for get expenses by tag
	router.GET("/", expensesHandler.GetExpenses)

	router.Run(":8282")

}
