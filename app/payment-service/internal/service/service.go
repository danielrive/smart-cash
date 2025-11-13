package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"smart-cash/payment-service/internal/common"
	"smart-cash/payment-service/internal/repositories"
	"smart-cash/payment-service/models"
	"time"

	"log/slog"

	"go.opentelemetry.io/otel"
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

func (s *PaymentService) ProcessPayment(ctx context.Context, paymentRequest models.PaymentRequest) (models.TransactionRequest, error) {
	// OTel instrumentation
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCProcessPayment")
	defer childSpan.End()

	s.logger.Info("processing payment",
		slog.String("user_id", paymentRequest.UserId),
		slog.String("expense_id", paymentRequest.ExpenseId),
		slog.String("component", "service"),
	)

	expense := models.Expense{}
	expenseBaseURL := fmt.Sprintf("http://expenses/expenses/%s", paymentRequest.ExpenseId)

	// Fetch expense details
	resp, err := http.Get(expenseBaseURL)
	if err != nil {
		s.logger.Error("error calling expense service",
			slog.String("error", err.Error()),
			slog.String("url", expenseBaseURL),
			slog.String("expense_id", paymentRequest.ExpenseId),
			slog.String("component", "service"),
		)
		return models.TransactionRequest{}, common.ErrExpenseNotFound
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("error reading response body from expense service",
			slog.String("error", err.Error()),
			slog.String("expense_id", paymentRequest.ExpenseId),
			slog.String("component", "service"),
		)
		return models.TransactionRequest{}, common.ErrInternalError
	}
	respBody, _ := io.ReadAll(resp.Body)

	err = json.Unmarshal(respBody, &expense)
	if err != nil {
		s.logger.Error("error parsing response body from expense service",
			slog.String("error", err.Error()),
			slog.String("expense_id", paymentRequest.ExpenseId),
			slog.String("component", "service"),
		)
		return models.TransactionRequest{}, common.ErrInternalError
	}
  // Validate if User exist and is not blocked
	// Validate if user exist
	if !s.validateUser(expense.UserId) {
		s.logger.Warn("user not found or not active",
			slog.String("user_id", expense.UserId),
			slog.String("expense_id", paymentRequest.ExpenseId),
			slog.String("component", "service"),
		)
		return models.TransactionRequest{}, common.ErrUserNotFound
	}

	// create transaction to bank

	transaction := models.TransactionRequest{
		TransactionId: s.uuid.New(),
		Date:          time.Now().UTC().Format("2006-01-02"),
		ExpenseId:     expense.ExpenseId,
		UserId:        expense.UserId,
		Amount:        expense.Amount,
		Status:        "pending",
	}

	err = s.paymentRepository.CreateTransaction(trContext, transaction)
	if err != nil {
		s.logger.Error("error creating transaction",
			slog.String("error", err.Error()),
			slog.String("transaction_id", transaction.TransactionId),
			slog.String("expense_id", paymentRequest.ExpenseId),
			slog.String("component", "service"),
		)
		transaction.Status = "notProcessed"
		return transaction, common.ErrInternalError
	}

	s.logger.Info("payment processed successfully",
		slog.String("transaction_id", transaction.TransactionId),
		slog.String("expense_id", paymentRequest.ExpenseId),
		slog.String("user_id", expense.UserId),
		slog.String("component", "service"),
	)

	return transaction, nil
}

func (s *PaymentService) GetTransaction(ctx context.Context, id string) (models.TransactionRequest, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCGetTransaction")
	defer childSpan.End()

	s.logger.Debug("getting transaction",
		slog.String("transaction_id", id),
		slog.String("component", "service"),
	)

	transaction, err := s.paymentRepository.GetTransaction(trContext, id)
	if err != nil {
		s.logger.Error("error getting transaction",
			slog.String("transaction_id", id),
			slog.String("error", err.Error()),
			slog.String("component", "service"),
		)
		return models.TransactionRequest{}, err
	}

	s.logger.Debug("transaction retrieved successfully",
		slog.String("transaction_id", id),
		slog.String("status", transaction.Status),
		slog.String("component", "service"),
	)

	return transaction, nil

}

func (s *PaymentService) validateUser(userId string) bool {
	userBaseURL := fmt.Sprintf("http://user/user/%s", userId)
	user := models.User{}

	// Validate if User exists and is not blocked
	resp, err := http.Get(userBaseURL)
	if err != nil {
		s.logger.Error("error calling user service",
			slog.String("error", err.Error()),
			slog.String("url", userBaseURL),
			slog.String("user_id", userId),
			slog.String("component", "service"),
		)
		return false
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("error reading response body from user service",
			slog.String("error", err.Error()),
			slog.String("url", userBaseURL),
			slog.String("user_id", userId),
			slog.String("component", "service"),
		)
		return false
	}

	err = json.Unmarshal(respBody, &user)
	if err != nil {
		s.logger.Error("error parsing response body from user service",
			slog.String("error", err.Error()),
			slog.String("url", userBaseURL),
			slog.String("user_id", userId),
			slog.String("component", "service"),
		)
		return false
	}

	s.logger.Debug("user validated",
		slog.String("user_id", userId),
		slog.Bool("active", user.Active),
		slog.String("component", "service"),
	)

	return user.Active
}
