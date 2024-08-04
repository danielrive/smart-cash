package main

import (
	"context"
	"log"
	"os"
	"smart-cash/expenses-service/internal/handler"
	"smart-cash/expenses-service/internal/repositories"
	"smart-cash/expenses-service/internal/service"
	"smart-cash/utils"

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
	// define uuid helper
	uuidHelper := utils.NewUUIDHelper()

	dynamoClient := dynamodb.NewFromConfig(cfg)
	// create a router with gin
	router := gin.New()
	router.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/health"),
		gin.Recovery(),
	)
	// // Initialize expenses repository
	expensesRepo := repositories.NewDynamoDBExpensesRepository(dynamoClient, expensesTable, uuidHelper)

	// Initialize expenses service
	expensesService := service.NewExpensesService(expensesRepo)

	// Init expenses handler
	expensesHandler := handler.NewExpensesHandler(expensesService)

	// create expenses
	router.POST("/expenses/", expensesHandler.CreateExpense)

	// define router for get expenses by tag
	router.GET("/expenses/:expenseId", expensesHandler.GetExpensesById)
	// define router for get expenses by category or userId
	router.GET("/expenses/", expensesHandler.GetExpensesByQuery)

	router.POST("/expenses/pay/", expensesHandler.PayExpenses)

	// Endpoint to test health check
	router.GET("/expenses/health", expensesHandler.HealthCheck)

	router.Run(":8282")

}
