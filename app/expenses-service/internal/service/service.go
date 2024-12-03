package service

import (
	"context"
	"log/slog"
	"smart-cash/expenses-service/internal/common"
	"smart-cash/expenses-service/internal/repositories"
	"smart-cash/expenses-service/models"
	"time"

	"go.opentelemetry.io/otel"
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

// Create a new expenses service
func NewExpensesService(expensesRepository *repositories.DynamoDBExpensesRepository, uuid UUIDHelper, logger *slog.Logger) *ExpensesService {
	return &ExpensesService{
		expensesRepository: expensesRepository,
		logger:             logger,
		uuid:               uuid,
	}
}

func (s *ExpensesService) CreateExpense(ctx context.Context, expense models.Expense) (models.ExpensesReturn, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCCreateExpense")
	childSpan.SetAttributes(attribute.String("component", "service"))
	defer childSpan.End()
	// set the expense status to unpaid
	expense.Status = "unpaid"
	// set the date of creation
	expense.Date = time.Now().UTC().Format("2006-01-02")
	// Create UUID
	expense.ExpenseId = s.uuid.New()

	if expense.Category == "" {
		expense.Category = "none"
	}
	response, err := s.expensesRepository.CreateExpense(trContext, expense)

	if err != nil {
		s.logger.Error("expense couldn't be created",
			"error", err.Error(),
		)
		return models.ExpensesReturn{}, err
	}
	return response, nil
}

// Function to get expenses by Id

func (s *ExpensesService) GetExpenseById(ctx context.Context, expenseId string) (models.Expense, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCGetExpenseById")
	childSpan.SetAttributes(attribute.String("component", "service"))
	defer childSpan.End()

	expense, err := s.expensesRepository.GetExpenseById(trContext, expenseId)
	if err != nil {
		return models.Expense{}, err
	}

	return expense, nil
}

// Delete expense

func (s *ExpensesService) DeleteExpense(ctx context.Context, expenseId string) (string, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCDeleteExpense")
	childSpan.SetAttributes(attribute.String("component", "service"))
	defer childSpan.End()

	expense, err := s.GetExpenseById(trContext, expenseId)
	if err != nil {
		return "", err
	}

	err = s.expensesRepository.DeleteExpenseById(trContext, expense.ExpenseId)
	if err != nil {
		return "", err
	}
	return expense.ExpenseId, nil
}

// Function to get expenses by userId or category

func (s *ExpensesService) GetExpByUserIdorCat(ctx context.Context, key string, value string) ([]models.Expense, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCGetExpByUserIdorCat")
	childSpan.SetAttributes(attribute.String("component", "service"))
	defer childSpan.End()

	expenses, err := s.expensesRepository.GetExpByUserIdorCat(trContext, key, value)
	if err != nil {
		return expenses, err
	}
	return expenses, nil
}
