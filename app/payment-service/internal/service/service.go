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
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCProcessPayment")
	defer childSpan.End()
	user := models.User{}
	expense := models.Expense{}
	//userBaseURL := fmt.Sprintf("http://127.0.0.1:8181/user/%s", paymentRequest.UserId)
	userBaseURL := fmt.Sprintf("http://user.%s/%s", common.DomainName, paymentRequest.UserId)
	//expenseBaseURL := fmt.Sprintf("http://127.0.0.1:8282/expenses/%s", paymentRequest.ExpenseId)
	expenseBaseURL := fmt.Sprintf("http://expenses.%s/%s", common.DomainName, paymentRequest.ExpenseId)

	// Validate if User exist and is not blocked
	resp, err := http.Get(userBaseURL)
	if err != nil {
		s.logger.Error("error creating the http request",
			"error", err.Error(),
			"url", userBaseURL,
		)
		return models.TransactionRequest{}, common.ErrUserNotFound
	}
	defer resp.Body.Close()

	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)

	err = json.Unmarshal(respBody, &user)
	if err != nil {
		s.logger.Error("error could not parse response body for user",
			"error", err.Error(),
		)
		return models.TransactionRequest{}, common.ErrInternalError
	}

	if !user.Active {
		return models.TransactionRequest{}, common.ErrUserBlocked
	}

	// Validate expense

	resp, err = http.Get(expenseBaseURL)
	if err != nil {
		s.logger.Error("error creating the http request",
			"error", err.Error(),
			"url", expenseBaseURL,
		)
		return models.TransactionRequest{}, common.ErrUserNotFound
	}
	respBody, _ = io.ReadAll(resp.Body)

	err = json.Unmarshal(respBody, &expense)
	if err != nil {
		s.logger.Error("error could not parse response body for expense",
			"error", err.Error(),
		)
		return models.TransactionRequest{}, common.ErrInternalError
	}

	// create transaction to bank

	transaction := models.TransactionRequest{
		TransactionId: s.uuid.New(),
		Date:          time.Now().UTC().Format("2006-01-02"),
		ExpenseId:     expense.ExpenseId,
		UserId:        user.UserId,
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
