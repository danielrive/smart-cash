package main

import (
	"context"
	"log"
	"os"

	"payment-service/internal/handler"
	"payment-service/internal/repositories"
	"payment-service/internal/service"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/gin-gonic/gin"
)

// Create a custom prometheus metric to count the numnber of request

func main() {
	// validate if env variables exists
	paymentTable := os.Getenv("DYNAMODB_PAYMENT_TABLE")
	if paymentTable == "" {
		panic("DYNAMODB_PAYMENT_TABLE cannot be empty")
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

	// init pyament repositories
	paymentRepository := repositories.NewDynamoDBPaymentRepository(dynamoClient, paymentTable)
	// init payment service
	paymentService := service.NewPaymentService(paymentRepository)
	// init payment handler
	paymentHandler := handler.NewPaymentHandler(paymentService)

	// GET api/v1[?userID=0&email(optinal)]
	router.GET("/", paymentHandler.GetOrder)

	// GET api/v1/[controller]/user[?userID=0]
	router.POST("/", paymentHandler.CreateOrder)

	router.Run(":8383")

}
