package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"slices"
	"syscall"
	"time"

	"smart-cash/user-service/internal/common"
	appconfig "smart-cash/user-service/internal/config"
	"smart-cash/user-service/internal/handler"
	"smart-cash/user-service/internal/handler/dto"
	"smart-cash/user-service/internal/repositories"
	"smart-cash/user-service/internal/service"
	"smart-cash/utils"
	"smart-cash/utils/logging"
	"smart-cash/utils/middleware"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel"
)

var (
	notToLogEndpoints = []string{"/user/health", "/user/metrics"}
)

func main() {
	if err := run(); err != nil {
		slog.Error("Application failed", "error", err)
		os.Exit(1)
	}
}

// run contains the main application logic and returns an error instead of exiting.
// This makes it testable and allows for better error handling.
func run() error {
	cfg, err := appconfig.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	common.ServiceName = cfg.ServiceName
	common.DomainName = cfg.DomainName

	tp := utils.InitOpenTelemetry(cfg.OtelCollector, cfg.ServiceName, cfg.Logger)
	otel.SetTracerProvider(tp)

	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithRegion(cfg.AwsRegion),
	)
	if err != nil {
		return fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	dynamoClient := dynamodb.NewFromConfig(awsCfg)
	uuidHelper := utils.NewUUIDHelper()

	router := gin.New()

	router.Use(
		otelgin.Middleware(cfg.ServiceName, otelgin.WithFilter(filterTraces)),
		logging.HTTPMiddleware(cfg.Logger, notToLogEndpoints),
		gin.Recovery(),
	)

	userRepo := repositories.NewDynamoDBUsersRepository(dynamoClient, cfg.UsersTable, uuidHelper, cfg.Logger)
	userService := service.NewUserService(userRepo, cfg.JWTSecret, cfg.Logger)
	userHandler := handler.NewUserHandler(userService, cfg.Logger)

	router.GET("/user/health", userHandler.HealthCheck)

	router.POST("/user",
		middleware.ValidateBody[dto.CreateUserRequest](),
		userHandler.CreateUser)

	router.POST("/user/login",
		middleware.ValidateBody[dto.LoginRequest](),
		userHandler.Login)

	router.GET("/user",
		middleware.RequireOneOfQueryParams([]string{"email", "username"}),
		userHandler.GetUserByQuery)

	router.GET("/user/:userId",
		middleware.AuthMiddleware(cfg.JWTSecret),
		middleware.ValidatePathParam("userId", "uuid"),
		userHandler.GetUserById)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:    ":8181",
		Handler: router,
	}

	errCh := make(chan error, 1)
	go func() {
		cfg.Logger.Info("starting server", "component", "main", "port", "8181")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return fmt.Errorf("server error: %w", err)
	case <-ctx.Done():
	}

	cfg.Logger.Info("shutdown signal received, shutting down gracefully...", "component", "main")

	if err := utils.ShutdownTracerProvider(context.Background(), tp, cfg.Logger); err != nil {
		cfg.Logger.Warn("tracer shutdown had issues, but continuing", "component", "main", "error", err)
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("server shutdown: %w", err)
	}

	cfg.Logger.Info("shutdown complete", "component", "main")
	return nil
}

func filterTraces(req *http.Request) bool {
	return slices.Index(notToLogEndpoints, req.URL.Path) == -1
}
