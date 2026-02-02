package consumers

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"smart-cash/payment-orchestrator-service/internal/common"
	"smart-cash/payment-orchestrator-service/models"
	"smart-cash/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type SQSConsumer struct {
	client     *sqs.Client
	queueURL   string
	logger     *slog.Logger
	processor  MessageProcessor
	maxWorkers int
}

// MessageProcessor defines the interface for processing SQS messages
type MessageProcessor interface {
	ProcessPaymentEvent(ctx context.Context, event models.PaymentEvent) error
}

// NewSQSConsumer creates a new SQS consumer
func NewSQSConsumer(queueURL string, awsRegion string, logger *slog.Logger, processor MessageProcessor) (*SQSConsumer, error) {
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion(awsRegion),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load AWS SDK config: %w", err)
	}

	return &SQSConsumer{
		client:     sqs.NewFromConfig(cfg),
		queueURL:   queueURL,
		logger:     logger,
		processor:  processor,
		maxWorkers: 10, // Configurable via env var if needed
	}, nil
}

// Start begins consuming messages from SQS
func (c *SQSConsumer) Start(ctx context.Context) error {
	c.logger.Info("starting SQS consumer",
		"queue_url", c.queueURL,
		"max_workers", c.maxWorkers,
	)

	// Use a worker pool pattern for concurrent message processing
	semaphore := make(chan struct{}, c.maxWorkers)

	for {
		// Check for shutdown before receiving
		if ctx.Err() != nil {
			c.logger.Info("SQS consumer stopping due to context cancellation")
			return ctx.Err()
		}

		// Receive messages in main loop (blocks for up to 10 seconds with long polling)
		// This prevents spawning goroutines when there are no messages
		messages, err := c.receiveMessages(ctx)
		if err != nil {
			// If context was cancelled during receive, exit
			if ctx.Err() != nil {
				c.logger.Info("SQS consumer stopping due to context cancellation")
				return ctx.Err()
			}
			c.logger.Error("failed to receive messages",
				"error", err,
			)
			// Continue to retry on other errors
			continue
		}

		// If no messages, loop again (no goroutine spawned)
		if len(messages) == 0 {
			continue
		}

		// Process each message in a separate goroutine (up to maxWorkers concurrent)
		for _, msg := range messages {
			// Check for shutdown before processing
			if ctx.Err() != nil {
				c.logger.Info("SQS consumer stopping, messages remaining in queue")
				return ctx.Err()
			}

			// Acquire semaphore (blocks if all workers are busy)
			semaphore <- struct{}{}

			// Process message in goroutine
			go func(m types.Message) {
				defer func() { <-semaphore }()

				if err := c.processMessage(ctx, m); err != nil {
					c.logger.Error("failed to process message",
						"error", err,
						"message_id", aws.ToString(m.MessageId),
					)
					// Message will remain in queue and be retried
					return
				}

				// Delete message only after successful processing
				if err := c.deleteMessage(ctx, m); err != nil {
					c.logger.Error("failed to delete message",
						"error", err,
						"message_id", aws.ToString(m.MessageId),
					)
				}
			}(msg)
		}
	}
}

// receiveMessages receives messages from SQS
func (c *SQSConsumer) receiveMessages(ctx context.Context) ([]types.Message, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "SQSConsumerReceiveMessages", "consumer",
		attribute.String("sqs.queue_url", c.queueURL),
	)
	defer endSpan()

	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.queueURL),
		MaxNumberOfMessages: 10, // SQS max is 10
		WaitTimeSeconds:     10, // Long polling
		VisibilityTimeout:   30, // 30 seconds to process
		MessageAttributeNames: []string{
			"All",
		},
	}

	resp, err := c.client.ReceiveMessage(ctx, input)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		return nil, fmt.Errorf("failed to receive messages: %w", err)
	}

	utils.AddSpanEvent(ctx, "messages received from SQS",
		attribute.Int("message.count", len(resp.Messages)),
	)

	return resp.Messages, nil
}

// processMessage processes a single SQS message
func (c *SQSConsumer) processMessage(ctx context.Context, msg types.Message) error {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "SQSConsumerProcessMessage", "consumer",
		attribute.String("sqs.message_id", aws.ToString(msg.MessageId)),
	)
	defer endSpan()

	var event models.PaymentEvent

	if err := json.Unmarshal([]byte(aws.ToString(msg.Body)), &event); err != nil {
		utils.RecordSpanError(ctx, err)
		return fmt.Errorf("failed to unmarshal message body: %w", err)
	}

	// Add payment event attributes to span
	utils.AddSpanEvent(ctx, "message unmarshaled",
		attribute.String("payment.id", event.PaymentId),
		attribute.String("expense.id", event.ExpenseId),
		attribute.String("user.id", event.UserId),
		attribute.Float64("payment.amount", event.Amount),
		attribute.String("payment.status", event.Status),
	)

	c.logger.Info("processing payment event",
		"payment_id", event.PaymentId,
		"expense_id", event.ExpenseId,
		"user_id", event.UserId,
		"amount", event.Amount,
		"status", event.Status,
	)

	// Process the payment event
	if err := c.processor.ProcessPaymentEvent(ctx, event); err != nil {
		utils.RecordSpanError(ctx, err)
		return fmt.Errorf("failed to process payment event: %w", err)
	}

	utils.SetSpanStatus(ctx, codes.Ok, "message processed successfully")
	return nil
}

// deleteMessage deletes a message from SQS after successful processing
func (c *SQSConsumer) deleteMessage(ctx context.Context, msg types.Message) error {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "SQSConsumerDeleteMessage", "consumer",
		attribute.String("sqs.message_id", aws.ToString(msg.MessageId)),
	)
	defer endSpan()

	input := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: msg.ReceiptHandle,
	}

	_, err := c.client.DeleteMessage(ctx, input)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		return fmt.Errorf("failed to delete message: %w", err)
	}

	utils.SetSpanStatus(ctx, codes.Ok, "message deleted successfully")

	c.logger.Debug("message deleted successfully",
		"message_id", aws.ToString(msg.MessageId),
	)

	return nil
}

// Stop gracefully stops the consumer
func (c *SQSConsumer) Stop(ctx context.Context) error {
	// Give workers time to finish processing
	timeout := time.After(30 * time.Second)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			c.logger.Warn("timeout waiting for workers to finish")
			return fmt.Errorf("timeout waiting for workers")
		case <-ticker.C:
			// Check if all workers are done (simplified - in production, track active workers)
			c.logger.Info("SQS consumer stopped")
			return nil
		}
	}
}
