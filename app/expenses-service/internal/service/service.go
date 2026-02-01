package service

import (
	"context"
	"log/slog"
	"smart-cash/expenses-service/internal/common"
	"smart-cash/expenses-service/internal/repositories"
	"smart-cash/expenses-service/models"
	"smart-cash/utils"
	"smart-cash/utils/logging"
	"time"

	"go.opentelemetry.io/otel/attribute"
)

// Define DynamoDB repository struct
type UUIDHelper interface {
	New() string
}

// Define service interface

type ExpensesService struct {
	expensesRepository *repositories.DynamoDBExpensesRepository
	logger             *slog.Logger
	uuid               UUIDHelper
}

// loggerWithTrace returns a logger with trace context if available in the context
func (s *ExpensesService) loggerWithTrace(ctx context.Context) *slog.Logger {
	return logging.LoggerWithTraceContext(ctx, s.logger)
}

// Create a new expenses service
func NewExpensesService(expensesRepository *repositories.DynamoDBExpensesRepository, uuid UUIDHelper, logger *slog.Logger) *ExpensesService {
	return &ExpensesService{
		expensesRepository: expensesRepository,
		logger:             logger,
		uuid:               uuid,
	}
}

func (s *ExpensesService) CreateExpense(ctx context.Context, expense models.Expense) (models.ExpensesReturn, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "CreateExpense", "service",
		attribute.String("user.id", expense.UserId),
		attribute.String("expense.name", expense.Name),
		attribute.Float64("expense.amount", expense.Amount),
	)
	defer endSpan()

	s.loggerWithTrace(ctx).Info("creating expense",
		slog.String("user_id", expense.UserId),
		slog.String("name", expense.Name),
		slog.Float64("amount", expense.Amount),
		slog.String("component", "service"),
	)

	// User validation is handled by JWT middleware - if user is inactive or doesn't exist,
	// they cannot authenticate and reach this point. The userId comes from JWT claims.
	// set the expense status to unpaid
	expense.Status = "unpaid"
	// set the date of creation
	expense.Date = time.Now().UTC()
	// Create UUID
	expense.ExpenseId = s.uuid.New()

	if expense.Category == "" {
		expense.Category = "none"
	}
	response, err := s.expensesRepository.CreateExpense(ctx, expense)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		s.loggerWithTrace(ctx).Error("expense couldn't be created",
			"error", err.Error(),
			"level", "service",
		)
		return models.ExpensesReturn{}, err
	}

	utils.AddSpanEvent(ctx, "expense created successfully",
		attribute.String("expense.id", response.ExpenseId),
		attribute.String("user.id", expense.UserId),
	)

	s.loggerWithTrace(ctx).Info("expense created successfully",
		slog.String("expense_id", response.ExpenseId),
		slog.String("user_id", expense.UserId),
		slog.String("component", "service"),
	)

	return response, nil
}

// Function to get expenses by Id

func (s *ExpensesService) GetExpenseById(ctx context.Context, expenseId string) (models.Expense, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "SVCGetExpenseById", "service",
		attribute.String("expense.id", expenseId),
	)
	defer endSpan()

	s.loggerWithTrace(ctx).Debug("getting expense by id",
		slog.String("expense_id", expenseId),
		slog.String("component", "service"),
	)

	expense, err := s.expensesRepository.GetExpenseById(ctx, expenseId)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		s.loggerWithTrace(ctx).Error("error getting expense by id",
			slog.String("expense_id", expenseId),
			slog.String("error", err.Error()),
			slog.String("component", "service"),
		)
		return models.Expense{}, err
	}

	utils.AddSpanEvent(ctx, "expense retrieved successfully",
		attribute.String("expense.id", expenseId),
	)

	s.loggerWithTrace(ctx).Debug("expense retrieved successfully",
		slog.String("expense_id", expenseId),
		slog.String("component", "service"),
	)

	return expense, nil
}

// Delete expense

func (s *ExpensesService) DeleteExpense(ctx context.Context, expenseId string) (string, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "SVCDeleteExpense", "service",
		attribute.String("expense.id", expenseId),
	)
	defer endSpan()

	s.loggerWithTrace(ctx).Info("deleting expense",
		slog.String("expense_id", expenseId),
		slog.String("component", "service"),
	)

	expense, err := s.GetExpenseById(ctx, expenseId)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		s.loggerWithTrace(ctx).Warn("expense not found for deletion",
			slog.String("expense_id", expenseId),
			slog.String("component", "service"),
		)
		return "", common.ErrExpenseNotFound
	}

	err = s.expensesRepository.DeleteExpenseById(ctx, expense.ExpenseId)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		s.loggerWithTrace(ctx).Error("error deleting expense",
			slog.String("expense_id", expenseId),
			slog.String("error", err.Error()),
			slog.String("component", "service"),
		)
		return "", err
	}

	utils.AddSpanEvent(ctx, "expense deleted successfully",
		attribute.String("expense.id", expenseId),
	)

	s.loggerWithTrace(ctx).Info("expense deleted successfully",
		slog.String("expense_id", expenseId),
		slog.String("component", "service"),
	)

	return expense.ExpenseId, nil
}

// Function to get expenses by userId or category

func (s *ExpensesService) GetExpByUserIdorCat(ctx context.Context, key string, value string) ([]models.Expense, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "SVCGetExpByUserIdorCat", "service",
		attribute.String("query.key", key),
		attribute.String("query.value", value),
	)
	defer endSpan()

	s.loggerWithTrace(ctx).Debug("getting expenses by user id or category",
		slog.String("key", key),
		slog.String("value", value),
		slog.String("component", "service"),
	)

	expenses, err := s.expensesRepository.GetExpByUserIdorCat(ctx, key, value)
	if err != nil {
		s.loggerWithTrace(ctx).Debug("expenses not found by query",
			slog.String("key", key),
			slog.String("value", value),
			slog.String("component", "service"),
		)
		return expenses, err
	}

	utils.AddSpanEvent(ctx, "expenses retrieved successfully",
		attribute.String("query.key", key),
		attribute.String("query.value", value),
		attribute.Int("expenses.count", len(expenses)),
	)

	s.loggerWithTrace(ctx).Debug("expenses retrieved successfully",
		slog.String("key", key),
		slog.Int("count", len(expenses)),
		slog.String("component", "service"),
	)

	return expenses, nil
}

// UpdateExpenseStatus updates the status of an expense
func (s *ExpensesService) UpdateExpenseStatus(ctx context.Context, expenseId string, status string) (models.ExpensesReturn, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "SVCUpdateExpenseStatus", "service",
		attribute.String("expense.id", expenseId),
		attribute.String("expense.status", status),
	)
	defer endSpan()

	s.loggerWithTrace(ctx).Info("updating expense status",
		slog.String("expense_id", expenseId),
		slog.String("status", status),
		slog.String("component", "service"),
	)

	// Get the expense first to ensure it exists
	expense, err := s.expensesRepository.GetExpenseById(ctx, expenseId)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		s.loggerWithTrace(ctx).Error("expense not found for status update",
			slog.String("expense_id", expenseId),
			slog.String("error", err.Error()),
			slog.String("component", "service"),
		)
		return models.ExpensesReturn{}, common.ErrExpenseNotFound
	}

	// Update the status
	expense.Status = status
	expense.UpdatedAt = time.Now().UTC()

	response, err := s.expensesRepository.UpdateExpenseStatus(ctx, expense)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		s.loggerWithTrace(ctx).Error("failed to update expense status",
			slog.String("expense_id", expenseId),
			slog.String("status", status),
			slog.String("error", err.Error()),
			slog.String("component", "service"),
		)
		return models.ExpensesReturn{}, err
	}

	utils.AddSpanEvent(ctx, "expense status updated successfully",
		attribute.String("expense.id", expenseId),
		attribute.String("expense.status", status),
	)

	s.loggerWithTrace(ctx).Info("expense status updated successfully",
		slog.String("expense_id", expenseId),
		slog.String("status", status),
		slog.String("component", "service"),
	)

	return response, nil
}
