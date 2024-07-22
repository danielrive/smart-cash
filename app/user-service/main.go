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

	router := gin.New()
	router.Use(
		gin.LoggerWithWriter(gin.DefaultWriter, "/health"),
		gin.Recovery(),
	)
	// new UUID helper
	uuidHelper := utils.NewUUIDHelper()

	// Initialize user repository
	userRepo := repositories.NewDynamoDBUsersRepository(dynamoClient, usersTable, uuidHelper)

	// Initialize user service
	userService := service.NewUserService(userRepo)

	// Init user handler
	userHandler := handler.NewUserHandler(userService)

	// GET user/userID
	router.GET("/user/:userId", userHandler.GetUserById)
	// GET user?username=username user?email=email
	router.GET("/user", userHandler.GetUserByQuery)

	// GET api/v1/[controller]/user[?userID=0]
	router.POST("/user", userHandler.CreateUser)

	// Health check
	router.GET("/user/health", userHandler.HealthCheck)

	router.Run(":8181")
}
