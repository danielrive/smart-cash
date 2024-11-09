package main

import (
	"context"
	"log/slog"
	"os"

	"smart-cash/user-service/internal/handler"
	"smart-cash/utils"

	"smart-cash/user-service/internal/repositories"
	"smart-cash/user-service/internal/service"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

var logger *slog.Logger

func main() {
	// Set-up Logger handler
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // (Info, Warn, Error)
	}))
	slog.SetDefault(logger)

	// Init OTel TracerProvider
	tp := initOpenTelemetry()

	otel.SetTracerProvider(tp)

	// validate if env variables exists
	usersTable := os.Getenv("DYNAMODB_USER_TABLE")
	if usersTable == "" {
		logger.Error("environment variable not found", slog.String("variable", "DYNAMODB_USER_TABLE"))
		os.Exit(1)
	}

	awsRegion := os.Getenv("AWS_REGION")
	if awsRegion == "" {
		logger.Error("environment variable not found", slog.String("variable", "AWS_REGION"))
		os.Exit(1)
	}

	// configure the SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
	)
	if err != nil {
		logger.Error("unable to load SDK config", slog.String("error", err.Error()))
	}
	dynamoClient := dynamodb.NewFromConfig(cfg)

	// create a router with gin

	router := gin.New()

	router.Use(otelgin.Middleware("user-service"))

	// logging for gin

	router.Use(gin.Recovery())

	// new UUID helper
	uuidHelper := utils.NewUUIDHelper()

	// Initialize user repository
	userRepo := repositories.NewDynamoDBUsersRepository(dynamoClient, usersTable, uuidHelper, logger)

	// Initialize user service
	userService := service.NewUserService(userRepo, logger)

	// Init user handler
	userHandler := handler.NewUserHandler(userService, logger)

	// GET user/userID
	router.GET("/user/:userId", userHandler.GetUserById)
	// GET user?username=username user?email=email
	//router.GET("/user", userHandler.GetUserByQuery)

	// GET api/v1/[controller]/user[?userID=0]
	router.POST("/user", userHandler.CreateUser)

	router.POST("/user/login", userHandler.Login)

	// Health check
	router.GET("/user/health", userHandler.HealthCheck)

	router.Run(":8181")
}
