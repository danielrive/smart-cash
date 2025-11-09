package service

import (
	"context"
	"log/slog"
	"smart-cash/expenses-service/internal/repositories"
	"smart-cash/expenses-service/models"

	"go.opentelemetry.io/otel"
)

// ExpenseService handles business logic for expenses
type ExpenseService struct {
	repo   *repositories.DynamoDBExpensesRepository
	logger *slog.Logger
}

// NewExpenseService creates a new expense service
func NewExpenseService(repo *repositories.DynamoDBExpensesRepository, logger *slog.Logger) *ExpenseService {
	return &ExpenseService{
		repo:   repo,
		logger: logger,
	}
}

// CreateExpense handles the creation of a new expense
func (s *ExpenseService) CreateExpense(ctx context.Context, expense models.Expense) (models.ExpensesReturn, error) {
	return s.repo.CreateExpense(ctx, expense)
}

// GetExpense retrieves an expense by ID
func (s *ExpenseService) GetExpense(ctx context.Context, expenseID string) (models.Expense, error) {
	return s.repo.GetExpense(ctx, expenseID)
}

// UpdateExpense updates an existing expense
func (s *ExpenseService) UpdateExpense(ctx context.Context, expense models.Expense) error {
	return s.repo.UpdateExpense(ctx, expense)
}

// DeleteExpense deletes an expense
func (s *ExpenseService) DeleteExpense(ctx context.Context, expenseID string) error {
	return s.repo.DeleteExpense(ctx, expenseID)
}

// ListExpenses retrieves a list of expenses with filters
func (s *ExpenseService) ListExpenses(ctx context.Context, userID string, filters repositories.ExpenseFilters) ([]models.Expense, error) {
	return s.repo.GetExpensesByDateRange(ctx, userID, filters)
}