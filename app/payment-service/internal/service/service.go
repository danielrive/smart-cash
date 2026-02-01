package service

import (
	"context"
	"smart-cash/payment-service/internal/common"
	"smart-cash/payment-service/internal/repositories"
	"smart-cash/payment-service/models"
	"smart-cash/utils"
	"smart-cash/utils/logging"
	"time"

	"log/slog"

	"go.opentelemetry.io/otel/attribute"
)

type UUIDHelper interface {
	New() string
}

// Define service interface

type PaymentService struct {
	paymentRepository *repositories.DynamoDBPaymentRepository
	logger            *slog.Logger
	uuid              UUIDHelper
}

// loggerWithTrace returns a logger with trace context if available in the context
func (s *PaymentService) loggerWithTrace(ctx context.Context) *slog.Logger {
	return logging.LoggerWithTraceContext(ctx, s.logger)
}

// Create a new Payment service
func NewPaymentService(paymentRepository *repositories.DynamoDBPaymentRepository, uuid UUIDHelper, logger *slog.Logger) *PaymentService {
	return &PaymentService{
		paymentRepository: paymentRepository,
		logger:            logger,
		uuid:              uuid,
	}
}

func (s *PaymentService) ProcessPayment(ctx context.Context, paymentRequest models.PaymentRequest) (models.PaymentResponse, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "SVCProcessPayment", "service",
		attribute.String("user.id", paymentRequest.UserId),
		attribute.String("expense.id", paymentRequest.ExpenseId),
	)
	defer endSpan()

	s.loggerWithTrace(ctx).Info("processing payment",
		slog.String("user_id", paymentRequest.UserId),
		slog.String("expense_id", paymentRequest.ExpenseId),
		slog.String("component", "service"),
	)

	// prepare Payment - DynamoDB stream will automatically publish to consumers
	request := models.PaymentRequest{
		PaymentId: s.uuid.New(),
		ExpenseId: paymentRequest.ExpenseId,
		UserId:    paymentRequest.UserId,
		Amount:    paymentRequest.Amount, // Assuming Amount is in paymentRequest
		Date:      time.Now().UTC(),
		Status:    models.PaymentStatusPending,
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	// Create Payment in DynamoDB
	// DynamoDB stream will automatically trigger Lambda which publishes to SQS
	response, err := s.paymentRepository.CreatePayment(ctx, request)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		s.loggerWithTrace(ctx).Error("Payment couldn't be created",
			slog.String("error", err.Error()),
			slog.String("level", "service"),
		)
		return models.PaymentResponse{}, err
	}

	utils.AddSpanEvent(ctx, "payment created successfully",
		attribute.String("payment.id", response.PaymentId),
		attribute.String("status", response.Status),
	)

	s.loggerWithTrace(ctx).Info("payment created successfully",
		slog.String("payment_id", response.PaymentId),
		slog.String("component", "service"),
	)

	return response, nil
}

func (s *PaymentService) GetPayment(ctx context.Context, id string) (models.PaymentRequest, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "SVCGetPayment", "service",
		attribute.String("payment.id", id),
	)
	defer endSpan()

	s.loggerWithTrace(ctx).Debug("getting payment",
		slog.String("payment_id", id),
		slog.String("component", "service"),
	)

	payment, err := s.paymentRepository.GetPayment(ctx, id)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		s.loggerWithTrace(ctx).Error("error getting payment",
			slog.String("payment_id", id),
			slog.String("error", err.Error()),
			slog.String("component", "service"),
		)
		return models.PaymentRequest{}, err
	}

	utils.AddSpanEvent(ctx, "payment retrieved successfully",
		attribute.String("payment.id", id),
		attribute.String("payment.status", payment.Status),
	)

	s.loggerWithTrace(ctx).Debug("payment retrieved successfully",
		slog.String("payment_id", id),
		slog.String("status", payment.Status),
		slog.String("component", "service"),
	)

	return payment, nil
}

// UpdatePaymentStatus updates the status of a payment
func (s *PaymentService) UpdatePaymentStatus(ctx context.Context, paymentId string, status string) error {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "SVCUpdatePaymentStatus", "service",
		attribute.String("payment.id", paymentId),
		attribute.String("payment.status", status),
	)
	defer endSpan()

	s.loggerWithTrace(ctx).Info("updating payment status",
		slog.String("payment_id", paymentId),
		slog.String("status", status),
		slog.String("component", "service"),
	)

	// Get the payment first to ensure it exists
	payment, err := s.paymentRepository.GetPayment(ctx, paymentId)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		s.loggerWithTrace(ctx).Error("payment not found for status update",
			slog.String("payment_id", paymentId),
			slog.String("error", err.Error()),
			slog.String("component", "service"),
		)
		return common.ErrPaymentNotFound
	}

	// Update the status
	payment.Status = status
	payment.UpdatedAt = time.Now().UTC()

	err = s.paymentRepository.UpdatePayment(ctx, payment)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		s.loggerWithTrace(ctx).Error("failed to update payment status",
			slog.String("payment_id", paymentId),
			slog.String("status", status),
			slog.String("error", err.Error()),
			slog.String("component", "service"),
		)
		return err
	}

	utils.AddSpanEvent(ctx, "payment status updated successfully",
		attribute.String("payment.id", paymentId),
		attribute.String("payment.status", status),
	)

	s.loggerWithTrace(ctx).Info("payment status updated successfully",
		slog.String("payment_id", paymentId),
		slog.String("status", status),
		slog.String("component", "service"),
	)

	return nil
}
