package main

import (
	"context"
	"log"
	"os"

	"user-service/internal/handler"

	"user-service/internal/repositories"
	"user-service/internal/service"

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

	// // Initialize user repository
	userRepo := repositories.NewDynamoDBUsersRepository(dynamoClient, usersTable)
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

	// GET api/v1/[controller]/user[?userID=0]
	router.Run(":8181")

}
