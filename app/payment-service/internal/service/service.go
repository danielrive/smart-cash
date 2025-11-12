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

	expense := models.Expense{}
	expenseBaseURL := fmt.Sprintf("http://expenses/expenses/%s", paymentRequest.ExpenseId)

	// Fetch expense details
	resp, err := http.Get(expenseBaseURL)
	if err != nil {
		s.logger.Error("error calling expense service",
			"error", err.Error(),
			"url", expenseBaseURL,
		)
		return models.TransactionRequest{}, common.ErrExpenseNotFound
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("error reading response body",
			"error", err.Error(),
		)
		return models.TransactionRequest{}, common.ErrInternalError
	}

	err = json.Unmarshal(respBody, &expense)
	if err != nil {
		s.logger.Error("error could not parse response body for expense",
			"error", err.Error(),
		)
		return models.TransactionRequest{}, common.ErrInternalError
	}

	// Validate if User exists and is not blocked
	if !s.validateUser(expense.UserId) {
		s.logger.Error("error user not found",
			"userId", expense.UserId,
			"level", "service",
		)
		return models.TransactionRequest{}, common.ErrUserNotFound
	}

	// Create transaction to bank
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
		s.logger.Error("error could not create the transaction",
			"error", err.Error(),
		)
		transaction.Status = "notProcessed"
		return transaction, common.ErrInternalError
	}

	return transaction, nil
}

func (s *PaymentService) GetTransaction(ctx context.Context, id string) (models.TransactionRequest, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCProcessPayment")
	defer childSpan.End()

	transaction, err := s.paymentRepository.GetTransaction(trContext, id)

	if err != nil {
		return models.TransactionRequest{}, err
	}

	return transaction, nil

}

func (s *PaymentService) validateUser(userId string) bool {
	userBaseURL := fmt.Sprintf("http://user/user/%s", userId)
	user := models.User{}

	// Validate if User exists and is not blocked
	resp, err := http.Get(userBaseURL)
	if err != nil {
		s.logger.Error("error calling user service",
			"error", err.Error(),
			"url", userBaseURL,
			"level", "service",
		)
		return false
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		s.logger.Error("error reading response body",
			"error", err.Error(),
			"level", "service",
		)
		return false
	}

	err = json.Unmarshal(respBody, &user)
	if err != nil {
		s.logger.Error("error could not parse response body for user",
			"error", err.Error(),
			"level", "service",
		)
		return false
	}

	return user.Active
}
