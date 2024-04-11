package main

import (
	"context"
	"log"
	"os"
	"smart-cash/expenses-service/internal/handler"
	"smart-cash/expenses-service/internal/repositories"
	"smart-cash/expenses-service/internal/service"
	"smart-cash/utils"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Create a custom prometheus metric to count the numnber of request

var (
	totalHttpRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Number of request made to the service",
		},
		[]string{"method", "path"},
	)

	// create metric to monitor the status code

	responseHttpStatus = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_response_status",
			Help: "Number of response status code",
		},
		[]string{"status", "path"},
	)

	// metric to count the time processing the request

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "http_request_duration",
			Help: "Duration of HTTP requests.",
		},
		[]string{"path"},
	)

	requestsMutex     sync.Mutex
	responseCodeMutex sync.Mutex
	durationMutex     sync.Mutex
)

// create gin middleware to count the requests

func prometheusMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// take the current time when the request arrives

		durationMutex.Lock()
		timer := prometheus.NewTimer(requestDuration.WithLabelValues(c.Request.URL.Path))
		defer func() {
			timer.ObserveDuration()
			durationMutex.Unlock()
		}()
		c.Next()
		// increase the metric for number of requests
		requestsMutex.Lock()
		totalHttpRequests.WithLabelValues(c.Request.Method, c.Request.URL.Path).Inc()
		requestsMutex.Unlock()

		// increase the metric for response status code
		responseCodeMutex.Lock()
		responseHttpStatus.WithLabelValues(strconv.Itoa(c.Writer.Status()), c.Request.URL.Path).Inc()
		responseCodeMutex.Unlock()

	}
}

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
	router := gin.Default()

	// using the middleware to collect http requests

	router.Use(prometheusMiddleware())

	// Creating route to monitoring app
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// // Initialize expenses repository
	expensesRepo := repositories.NewDynamoDBExpensesRepository(dynamoClient, expensesTable, uuidHelper)

	// Initialize expenses service
	expensesService := service.NewExpensesService(expensesRepo)

	// Init expenses handler
	expensesHandler := handler.NewExpensesHandler(expensesService)

	router.POST("/", expensesHandler.CreateExpense)

	//router.GET("/calculateTotal", expensesHandler.CalculateTotalPerCategory)

	// define router for get expenses by tag
	router.GET("/", expensesHandler.GetExpenses)

	router.GET("/health", expensesHandler.HealthCheck)

	router.Run(":8282")

}
