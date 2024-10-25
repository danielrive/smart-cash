package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	bank "smart-cash/bank-service/models"
	expense "smart-cash/expenses-service/models"
	user "smart-cash/user-service/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type modelsStructs interface {
	bank.BankUser | expense.Expense | user.User
}

var Logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
	Level: slog.LevelDebug, // (Info, Warn, Error)
}))

func main() {
	// services in project

	services := [3]string{"bank", "user", "expenses"}

	awsRegion := os.Getenv("AWS_REGION")

	if awsRegion == "" {
		Logger.Error("environment variable not found", slog.String("variable", "AWS_REGION"))
		os.Exit(1)
	}

	// configure the SDK
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(awsRegion),
	)
	if err != nil {
		Logger.Error("unable to load SDK config", slog.String("error", err.Error()))
	}

	s3Client := s3.NewFromConfig(cfg)

	dynamoClient := dynamodb.NewFromConfig(cfg)

	// Get files

	for _, svc := range services {
		data := getData[expense.Expense]("smart-cash-fake-data", svc+"_service_HYDRA.json", s3Client)
		hydraDynamodb(svc+"-table", dynamoClient, data)
	}

}

func getData[T modelsStructs](bucket string, fileName string, s3Client *s3.Client) []T {
	// Input file

	input := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
	}

	output, err := s3Client.GetObject(context.TODO(), input)

	if err != nil {
		Logger.Error("unable to download file",
			"error", err.Error(),
			"filename", fileName,
		)
		return nil
	}
	defer output.Body.Close()

	fileContent, err := io.ReadAll(output.Body)

	if err != nil {
		Logger.Error("unable to create local file",
			"error", err.Error(),
			"filename", fileName,
		)
		return nil
	}

	var serviceStruct []T

	fmt.Printf("File downloaded successfully to %s\n", fileName)

	_ = json.Unmarshal(fileContent, &serviceStruct)
	return serviceStruct

}

// func hydraDynamodb[T modelsProject](tableName string, dynamoClient *dynamodb.Client,data []byte, svcStruct T ) {
func hydraDynamodb[T modelsStructs](tableName string, dynamoClient *dynamodb.Client, data []T) {

	for _, object := range data {

		item, err := attributevalue.MarshalMap(object)

		if err != nil {
			Logger.Error("error while unmarshaling DynamoDB item",
				"error", err.Error(),
				"data", object,
				"table", tableName,
			)
		}

		_, err = dynamoClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
			TableName: aws.String(tableName),
			Item:      item,
		})

		if err != nil {
			Logger.Error("dynamodb error while putting item",
				"error", err.Error(),
				"data", object,
				"table", tableName,
			)
		}

	}
	Logger.Info("dynamodb operation finished",
		"table", tableName,
	)

}
