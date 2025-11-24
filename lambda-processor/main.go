package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

var (
	logger    *slog.Logger
	sqsClient *sqs.Client
	queueURL  string
)

// TransactionEvent represents a transaction from DynamoDB stream
type TransactionEvent struct {
	TransactionId string    `json:"transactionId"`
	UserId        string    `json:"userId"`
	ExpenseId     string    `json:"expenseId"`
	Amount        float64   `json:"amount"`
	Date          time.Time `json:"date"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// PublishedEvent is what gets sent to SQS
type PublishedEvent struct {
	EventType string           `json:"eventType"` // "TransactionCreated"
	Timestamp time.Time        `json:"timestamp"`
	Data      TransactionEvent `json:"data"`
}

// OutboxEvent represents the structure from DynamoDB
type OutboxEvent struct {
	EventID       string    `dynamodbav:"eventId"`
	AggregateID   string    `dynamodbav:"aggregateId"`
	AggregateType string    `dynamodbav:"aggregateType"`
	EventType     string    `dynamodbav:"eventType"`
	Payload       string    `dynamodbav:"payload"`
	Status        string    `dynamodbav:"status"`
	CreatedAt     time.Time `dynamodbav:"createdAt"`
	Attempt       int       `dynamodbav:"attempt"`
	LastError     *string   `dynamodbav:"lastError"`
}

func init() {
	// Initialize logger
	logger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
	slog.SetDefault(logger)

	// Load configuration from environment
	queueURL = os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		logger.Error("SQS_QUEUE_URL not set")
		os.Exit(1)
	}

	// Initialize AWS clients
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		logger.Error("failed to load AWS config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	sqsClient = sqs.NewFromConfig(cfg)

	logger.Info("Lambda DynamoDB Stream processor initialized",
		slog.String("queue_url", queueURL),
	)
}

func main() {
	lambda.Start(handleDynamoDBStream)
}

// handleDynamoDBStream processes DynamoDB stream records and publishes to SQS
func handleDynamoDBStream(ctx context.Context, event events.DynamoDBEvent) error {
	logger.Info("processing DynamoDB stream event",
		slog.Int("record_count", len(event.Records)),
	)

	successCount := 0
	failureCount := 0
	var lastError error

	for _, record := range event.Records {
		logger.Debug("processing record",
			slog.String("event_id", record.EventID),
			slog.String("event_name", record.EventName),
		)

		// Only process INSERT and MODIFY events (ignore REMOVE)
		if record.EventName != events.DynamoDBOperationTypeInsert && record.EventName != events.DynamoDBOperationTypeModify {
			logger.Debug("skipping record", slog.String("event_name", record.EventName))
			continue
		}

		// Extract transaction from DynamoDB stream image
		transaction, err := extractTransaction(record.DynamoDB.NewImage)
		if err != nil {
			logger.Error("failed to extract transaction",
				slog.String("error", err.Error()),
				slog.String("event_id", record.EventID),
			)
			failureCount++
			lastError = err
			continue
		}

		// Publish to SQS
		err = publishToSQS(ctx, transaction)
		if err != nil {
			logger.Error("failed to publish to SQS",
				slog.String("error", err.Error()),
				slog.String("transaction_id", transaction.TransactionId),
			)
			failureCount++
			lastError = err
			continue
		}

		logger.Info("successfully published transaction to SQS",
			slog.String("transaction_id", transaction.TransactionId),
			slog.String("event_name", record.EventName),
		)
		successCount++
	}

	logger.Info("stream processing completed",
		slog.Int("success_count", successCount),
		slog.Int("failure_count", failureCount),
	)

	// Return error if any processing failed, Lambda will retry
	if failureCount > 0 {
		return fmt.Errorf("failed to process %d events: %w", failureCount, lastError)
	}

	return nil
}

// extractTransaction extracts transaction data from DynamoDB stream image
func extractTransaction(image map[string]events.DynamoDBAttributeValue) (*TransactionEvent, error) {
	var transaction TransactionEvent

	// Extract each field from the DynamoDB image
	if val, exists := image["transactionId"]; exists {
		if len(val.S) > 0 {
			transaction.TransactionId = val.S
		}
	}

	if val, exists := image["userId"]; exists {
		if len(val.S) > 0 {
			transaction.UserId = val.S
		}
	}

	if val, exists := image["expenseId"]; exists {
		if len(val.S) > 0 {
			transaction.ExpenseId = val.S
		}
	}

	if val, exists := image["amount"]; exists {
		// DynamoDB stores numbers as strings in streams
		if len(val.N) > 0 {
			fmt.Sscanf(val.N, "%f", &transaction.Amount)
		}
	}

	if val, exists := image["date"]; exists {
		if len(val.S) > 0 {
			transaction.Date, _ = time.Parse(time.RFC3339, val.S)
		}
	}

	if val, exists := image["status"]; exists {
		if len(val.S) > 0 {
			transaction.Status = val.S
		}
	}

	if val, exists := image["createdAt"]; exists {
		if len(val.S) > 0 {
			transaction.CreatedAt, _ = time.Parse(time.RFC3339, val.S)
		}
	}

	if val, exists := image["updatedAt"]; exists {
		if len(val.S) > 0 {
			transaction.UpdatedAt, _ = time.Parse(time.RFC3339, val.S)
		}
	}

	if transaction.TransactionId == "" {
		return nil, fmt.Errorf("transaction ID not found in stream image")
	}

	return &transaction, nil
}

// publishToSQS publishes the transaction event to SQS
func publishToSQS(ctx context.Context, transaction *TransactionEvent) error {
	// Create the published event
	publishedEvent := PublishedEvent{
		EventType: "TransactionCreated",
		Timestamp: time.Now().UTC(),
		Data:      *transaction,
	}

	messageBody, err := json.Marshal(publishedEvent)
	if err != nil {
		return fmt.Errorf("error marshaling event: %w", err)
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(queueURL),
		MessageBody: aws.String(string(messageBody)),
		MessageAttributes: map[string]sqstypes.MessageAttributeValue{
			"EventType": {
				StringValue: aws.String("TransactionCreated"),
				DataType:    aws.String("String"),
			},
			"TransactionId": {
				StringValue: aws.String(transaction.TransactionId),
				DataType:    aws.String("String"),
			},
			"UserId": {
				StringValue: aws.String(transaction.UserId),
				DataType:    aws.String("String"),
			},
		},
	}

	// If using FIFO queue, add deduplication ID and message group ID
	if isFIFOQueue(queueURL) {
		input.MessageDeduplicationId = aws.String(transaction.TransactionId)
		input.MessageGroupId = aws.String(transaction.UserId) // Group by user ID
	}

	result, err := sqsClient.SendMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("error sending message to SQS: %w", err)
	}

	logger.Debug("message sent to SQS",
		slog.String("transaction_id", transaction.TransactionId),
		slog.String("message_id", *result.MessageId),
	)

	return nil
}

// isFIFOQueue checks if the queue URL is for a FIFO queue
func isFIFOQueue(queueURL string) bool {
	return strings.Contains(queueURL, ".fifo")
}
