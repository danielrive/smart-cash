package service

import (
	"context"
	"smart-cash/bank-service/internal/common"
	"smart-cash/bank-service/internal/repositories"
	"smart-cash/bank-service/models"

	"log/slog"

	"go.opentelemetry.io/otel"
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

func (s *BankService) ProcessPayment(ctx context.Context, transaction models.TransactionRequest) (models.TransactionRequest, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCProcessPayment")
	defer childSpan.End()

	s.logger.Info("processing bank payment",
		slog.String("transaction_id", transaction.TransactionId),
		slog.String("expense_id", transaction.ExpenseId),
		slog.String("user_id", transaction.UserId),
		slog.Float64("amount", transaction.Amount),
		slog.String("component", "service"),
	)

	user, err := s.bankRepository.GetUser(trContext, transaction.UserId)
	if err != nil {
		s.logger.Error("failed getting user for payment",
			slog.String("user_id", transaction.UserId),
			slog.String("error", err.Error()),
			slog.String("component", "service"),
		)
		transaction.Status = "NotPaid"
		return transaction, common.ErrTransactionFailed
	}
	
	newSaldo, err := processPayment(transaction.Amount, user.Savings)
	if err != nil {
		s.logger.Warn("transaction failed - insufficient funds",
			slog.String("error", err.Error()),
			slog.String("user_id", transaction.UserId),
			slog.String("expense_id", transaction.ExpenseId),
			slog.Float64("amount", transaction.Amount),
			slog.Float64("current_savings", user.Savings),
			slog.String("component", "service"),
		)
		transaction.Status = "NotPaid"
		return transaction, common.ErrTransactionFailed
	}
	
	// update saving in user account
	user.Savings = newSaldo
	err = s.bankRepository.UpdateSavingsUser(trContext, user)
	if err != nil {
		s.logger.Error("transaction failed - could not update savings",
			slog.String("error", err.Error()),
			slog.String("user_id", transaction.UserId),
			slog.String("expense_id", transaction.ExpenseId),
			slog.String("component", "service"),
		)
		transaction.Status = "NotPaid"
		return transaction, err
	}
	
	transaction.Status = "Paid"
	s.logger.Info("bank payment processed successfully",
		slog.String("transaction_id", transaction.TransactionId),
		slog.String("user_id", transaction.UserId),
		slog.String("expense_id", transaction.ExpenseId),
		slog.Float64("new_savings", newSaldo),
		slog.String("component", "service"),
	)

	return transaction, nil
}

// Function to get bank by Id
func (s *BankService) GetUser(ctx context.Context, userId string) (models.BankUser, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCGetUser")
	defer childSpan.End()

	s.logger.Debug("getting bank user",
		slog.String("user_id", userId),
		slog.String("component", "service"),
	)

	user, err := s.bankRepository.GetUser(trContext, userId)
	if err != nil {
		s.logger.Error("error getting bank user",
			slog.String("error", err.Error()),
			slog.String("user_id", userId),
			slog.String("component", "service"),
		)
		return models.BankUser{}, err
	}

	if user.Blocked {
		s.logger.Warn("user is blocked - transactions not allowed",
			slog.String("user_id", userId),
			slog.String("component", "service"),
		)
		return models.BankUser{}, common.ErrUserBlocked
	}

	s.logger.Debug("bank user retrieved successfully",
		slog.String("user_id", userId),
		slog.String("currency", user.Currency),
		slog.String("component", "service"),
	)

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
