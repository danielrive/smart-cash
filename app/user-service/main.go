package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"slices"

	"smart-cash/user-service/internal/common"
	"smart-cash/user-service/internal/handler"
	"smart-cash/user-service/internal/handler/dto"
	"smart-cash/user-service/internal/repositories"
	"smart-cash/user-service/internal/service"
	"smart-cash/utils"
	"smart-cash/utils/logging"
	"smart-cash/utils/middleware"

	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

var (
	otelCollector     string
	usersTable        string
	awsRegion         string
	jwtSecret         []byte
	notToLogEndpoints = []string{"/user/health", "/user/metrics"}
	logger            *slog.Logger
	domainName        string
)

func init() {
	// Logger config
	logsConfig := logging.LoadConfig()
	logger = logging.InitLogger(logsConfig, common.ServiceName)
	// validate ENV variables
	common.DomainName = os.Getenv("DOMAIN_NAME")
	if domainName == "" {
		common.DomainName = "localhost"
	}

	usersTable = os.Getenv("DYNAMODB_USER_TABLE")
	if usersTable == "" {
		logger.Error("environment variable not found", slog.String("variable", "DYNAMODB_USER_TABLE"))
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
	if common.ServiceName == "" {
		logger.Error("environment variable not found", slog.String("variable", "SERVICE_NAME"))
		os.Exit(1)
	}

	jwtSecretStr := os.Getenv("JWT_SECRET")
	if jwtSecretStr == "" {
		logger.Warn("JWT_SECRET not set, using default (NOT FOR PRODUCTION!)")
		jwtSecretStr = "default-secret-change-me"
	}
	jwtSecret = []byte(jwtSecretStr)

}

func main() {
	// Init OTel TracerProvider
	tp := utils.InitOpenTelemetry(otelCollector, common.ServiceName, logger)

	otel.SetTracerProvider(tp)

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

	router.Use(
		otelgin.Middleware(common.ServiceName, otelgin.WithFilter(filterTraces)),
		logging.HTTPMiddleware(logger, notToLogEndpoints), // Add structured logging middleware
		gin.Recovery(),
	)

	// new UUID helper
	uuidHelper := utils.NewUUIDHelper()

	// Initialize user repository
	userRepo := repositories.NewDynamoDBUsersRepository(dynamoClient, usersTable, uuidHelper, logger)

	// Initialize user service (pass JWT secret!)
	userService := service.NewUserService(userRepo, jwtSecret, logger)

	// Init user handler
	userHandler := handler.NewUserHandler(userService, logger)

	// Public routes
	router.GET("/user/health", userHandler.HealthCheck)

	router.POST("/user",
		middleware.ValidateBody[dto.CreateUserRequest](), // Validate user creation
		userHandler.CreateUser)

	router.POST("/user/login",
		middleware.ValidateBody[dto.LoginRequest](), // Validate login
		userHandler.Login)

	router.GET("/user",
		middleware.RequireOneOfQueryParams([]string{"email", "username"}), // Require email OR username
		userHandler.GetUserByQuery)

	// Protected routes
	router.GET("/user/:userId",
		middleware.AuthMiddleware(jwtSecret),
		middleware.ValidatePathParam("userId", "uuid"), // Validate userId is UUID
		userHandler.GetUserById)

	router.Run(":8181")
}

func filterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}
