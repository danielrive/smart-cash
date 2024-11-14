package service

import (
	"context"
	"smart-cash/bank-service/internal/common"
	"smart-cash/bank-service/internal/repositories"
	"smart-cash/bank-service/models"

	"log/slog"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
)

// Define service interface

type BankService struct {
	bankRepository *repositories.DynamoDBBankRepository
	logger         *slog.Logger
}

// Create a new bank service
func NewBankService(bankRepository *repositories.DynamoDBBankRepository, logger *slog.Logger) *BankService {
	return &BankService{
		bankRepository: bankRepository,
		logger:         logger,
	}
}

func (s *BankService) ProcessPayment(ctx context.Context, transaction models.PaymentRequest) (models.PaymentRequest, error) {
	tr := otel.Tracer("bank-service")
	trContext, childSpan := tr.Start(ctx, "SVCProcessPayment")
	childSpan.SetAttributes(attribute.String("component", "service"))
	defer childSpan.End()

	// validate user in bank
	user, err := s.GetUser(trContext, transaction.UserId)
	if err != nil || user.Blocked {
		s.logger.Error("payment can not be processed",
			"error", err.Error(),
			"userId", transaction.UserId,
		)
		transaction.Status = "NotPaid"
		return transaction, common.ErrTransactionFailed
	}

	s.logger.Info("processing transaction for expense",
		"expenseId", transaction.ExpenseId,
		"userId", transaction.UserId,
	)
	newSaldo, err := processPayment(transaction.Amount, user.Savings)
	if err != nil {
		s.logger.Error("transaction failed",
			"error", err.Error(),
			"userId", transaction.UserId,
			"expenseId", transaction.ExpenseId,
		)
		transaction.Status = "NotPaid"
		return transaction, common.ErrTransactionFailed
	}
	// update saving in user account
	user.Savings = newSaldo
	err = s.bankRepository.UpdateSavingsUser(trContext, user)
	if err != nil {
		// Use retry ?
		s.logger.Error("transaction failed",
			"error", err.Error(),
			"userId", transaction.UserId,
			"expenseId", transaction.ExpenseId,
		)
		transaction.Status = "NotPaid"
		return transaction, err
	}
	transaction.Status = "Paid"
	s.logger.Info("transaction processed",
		"userId", transaction.UserId,
		"expenseId", transaction.ExpenseId,
	)

	return transaction, nil
}

// Function to get bank by Id
func (s *BankService) GetUser(ctx context.Context, userId string) (models.BankUser, error) {
	tr := otel.Tracer("bank-service")
	trContext, childSpan := tr.Start(ctx, "SVCGetUser")
	childSpan.SetAttributes(attribute.String("component", "service"))
	defer childSpan.End()

	user, err := s.bankRepository.GetUser(trContext, userId)

	if err != nil {
		s.logger.Error("error getting the user",
			"error", err.Error(),
			"user", userId,
		)
		return models.BankUser{}, err
	}

	if user.Blocked {
		s.logger.Error("not transactions allowed",
			"error", common.ErrUserBlocked,
			"user", userId,
		)
		return models.BankUser{}, common.ErrUserBlocked
	}

	return user, nil

}

func processPayment(amount, savings float64) (float64, error) {
	newSavings := savings - amount
	// Check if the new savings is negative
	if newSavings < 0 {
		return 0, common.ErrInsufficientFundsError
	}
	return newSavings, nil
}
