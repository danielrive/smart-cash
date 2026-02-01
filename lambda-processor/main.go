package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	sqstypes "github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type PaymentEvent struct {
	PaymentId string  `json:"paymentId"`
	UserId    string  `json:"userId"`
	ExpenseId string  `json:"expenseId"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

type Handler struct {
	sqsClient *sqs.Client
	queueURL  string
	logger    *slog.Logger
	isFIFO    bool
}

func NewHandler(ctx context.Context) (*Handler, error) {
	qURL := os.Getenv("SQS_QUEUE_URL")
	if qURL == "" {
		return nil, fmt.Errorf("SQS_QUEUE_URL environment variable is required")
	}

	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	return &Handler{
		sqsClient: sqs.NewFromConfig(cfg),
		queueURL:  qURL,
		logger:    slog.New(slog.NewJSONHandler(os.Stdout, nil)),
		isFIFO:    strings.HasSuffix(qURL, ".fifo"),
	}, nil
}

func (h *Handler) Handle(ctx context.Context, event events.DynamoDBEvent) (events.DynamoDBEventResponse, error) {
	var batchFailures []events.DynamoDBBatchItemFailure

	for _, record := range event.Records {
		// Only process NEW data (INSERT) or UPDATED data (MODIFY)
		if record.EventName == "REMOVE" {
			continue
		}

		err := h.processRecord(ctx, record)
		if err != nil {
			h.logger.Error("failed to process record",
				slog.String("error", err.Error()),
				slog.String("eventID", record.EventID),
			)
			// Reporting the SequenceNumber tells Lambda to retry ONLY this specific message
			batchFailures = append(batchFailures, events.DynamoDBBatchItemFailure{
				ItemIdentifier: record.Change.SequenceNumber,
			})
		}
	}

	return events.DynamoDBEventResponse{BatchItemFailures: batchFailures}, nil
}

func (h *Handler) processRecord(ctx context.Context, record events.DynamoDBEventRecord) error {
	image := record.Change.NewImage

	payment := PaymentEvent{
		PaymentId: image["paymentId"].String(),
		UserId:    image["userId"].String(),
		ExpenseId: image["expenseId"].String(),
		Status:    image["status"].String(),
	}

	// Handle numeric conversion
	amountStr := image["amount"].Number()
	if amountStr != "" {
		fmt.Sscanf(amountStr, "%f", &payment.Amount)
	}

	if payment.PaymentId == "" {
		return fmt.Errorf("paymentId is missing in stream record")
	}

	return h.sendToSQS(ctx, payment)
}

func (h *Handler) sendToSQS(ctx context.Context, payment PaymentEvent) error {
	body, err := json.Marshal(payment)
	if err != nil {
		return fmt.Errorf("failed to marshal payment: %w", err)
	}

	input := &sqs.SendMessageInput{
		QueueUrl:    aws.String(h.queueURL),
		MessageBody: aws.String(string(body)),
		MessageAttributes: map[string]sqstypes.MessageAttributeValue{
			"PaymentId": {DataType: aws.String("String"), StringValue: aws.String(payment.PaymentId)},
			"UserId":    {DataType: aws.String("String"), StringValue: aws.String(payment.UserId)},
		},
	}

	// Handle FIFO logic if necessary
	if h.isFIFO {
		input.MessageDeduplicationId = aws.String(payment.PaymentId)
		input.MessageGroupId = aws.String(payment.UserId)
	}

	_, err = h.sqsClient.SendMessage(ctx, input)
	if err != nil {
		return fmt.Errorf("SQS send error: %w", err)
	}

	h.logger.Info("published to SQS", slog.String("paymentId", payment.PaymentId))
	return nil
}

func main() {
	ctx := context.Background()

	h, err := NewHandler(ctx)
	if err != nil {
		slog.Error("initialization failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	lambda.Start(h.Handle)
}
