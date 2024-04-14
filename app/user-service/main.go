package main

import (
	"context"
	"log"
	"os"

	"smart-cash/user-service/internal/handler"
	"smart-cash/utils"

	"smart-cash/user-service/internal/repositories"
	"smart-cash/user-service/internal/service"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
)

func main() {
	// validate if env variables exists
	usersTable := os.Getenv("DYNAMODB_USER_TABLE")
	if usersTable == "" {
		panic("DYNAMODB_USER_TABLE cannot be empty")
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
	// new UUID helper
	uuidHelper := utils.NewUUIDHelper()

	// // Initialize user repository
	userRepo := repositories.NewDynamoDBUsersRepository(dynamoClient, usersTable, uuidHelper)
	// Initialize user service
	userService := service.NewUserService(userRepo)

	// Init user handler
	userHandler := handler.NewUserHandler(userService)

	// GET api/v1[?userID=0&email(optinal)]
	router.GET("/", userHandler.GetUser)

	// GET api/v1/[controller]/user[?userID=0]
	router.POST("/", userHandler.CreateUser)

	// login method, will return a token and userId
	router.POST("/login", userHandler.Login)

	// Health check
	router.GET("/health", userHandler.HealthCheck)

	// test connect to other services

	router.GET("/connectToSvc", userHandler.ConnectToOtherSvc)

	// GET api/v1/[controller]/user[?userID=0]
	router.Run(":8181")
	// Find User by email, userId and username
}
