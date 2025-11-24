package service

import (
	"context"
	"smart-cash/payment-service/internal/common"
	"smart-cash/payment-service/internal/repositories"
	"smart-cash/payment-service/models"
	"smart-cash/utils"
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

	s.logger.Info("processing payment",
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
		s.logger.Error("Payment couldn't be created",
			slog.String("error", err.Error()),
			slog.String("level", "service"),
		)
		return models.PaymentResponse{}, err
	}

	utils.AddSpanEvent(ctx, "payment created successfully",
		attribute.String("payment.id", response.PaymentId),
		attribute.String("status", response.Status),
	)

	s.logger.Info("payment created successfully",
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

	s.logger.Debug("getting payment",
		slog.String("payment_id", id),
		slog.String("component", "service"),
	)

	payment, err := s.paymentRepository.GetPayment(ctx, id)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		s.logger.Error("error getting payment",
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

	s.logger.Debug("payment retrieved successfully",
		slog.String("payment_id", id),
		slog.String("status", payment.Status),
		slog.String("component", "service"),
	)

	return payment, nil

}
