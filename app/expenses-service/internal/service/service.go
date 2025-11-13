package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
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
	// OTel trace instrumentation
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "CreateExpense")
	childSpan.SetAttributes(attribute.String("component", "service"))
	defer childSpan.End()

	s.logger.Info("creating expense",
		slog.String("user_id", expense.UserId),
		slog.String("name", expense.Name),
		slog.Float64("amount", expense.Amount),
		slog.String("component", "service"),
	)

	// Validate if user exist
	if !s.validateUser(expense.UserId) {
		s.logger.Warn("user not found during expense creation",
			slog.String("user_id", expense.UserId),
			slog.String("component", "service"),
		)
		return models.ExpensesReturn{}, common.ErrUserNotFound
	}
	// set the expense status to unpaid
	expense.Status = "unpaid"
	// set the date of creation
	expense.Date = time.Now().UTC()
	// Create UUID
	expense.ExpenseId = s.uuid.New()

	if expense.Category == "" {
		expense.Category = "none"
	}
	response, err := s.expensesRepository.CreateExpense(trContext, expense)
	if err != nil {
		s.logger.Error("expense couldn't be created",
			slog.String("error", err.Error()),
			slog.String("expense_id", expense.ExpenseId),
			slog.String("user_id", expense.UserId),
			slog.String("component", "service"),
		)
		return models.ExpensesReturn{}, err
	}

	s.logger.Info("expense created successfully",
		slog.String("expense_id", response.ExpenseId),
		slog.String("user_id", expense.UserId),
		slog.String("component", "service"),
	)

	return response, nil
}

// Function to get expenses by Id

func (s *ExpensesService) GetExpenseById(ctx context.Context, expenseId string) (models.Expense, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCGetExpenseById")
	defer childSpan.End()

	s.logger.Debug("getting expense by id",
		slog.String("expense_id", expenseId),
		slog.String("component", "service"),
	)

	expense, err := s.expensesRepository.GetExpenseById(trContext, expenseId)
	if err != nil {
		s.logger.Error("error getting expense by id",
			slog.String("expense_id", expenseId),
			slog.String("error", err.Error()),
			slog.String("component", "service"),
		)
		return models.Expense{}, err
	}

	s.logger.Debug("expense retrieved successfully",
		slog.String("expense_id", expenseId),
		slog.String("component", "service"),
	)

	return expense, nil
}

// Delete expense

func (s *ExpensesService) DeleteExpense(ctx context.Context, expenseId string) (string, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCDeleteExpense")
	defer childSpan.End()

	s.logger.Info("deleting expense",
		slog.String("expense_id", expenseId),
		slog.String("component", "service"),
	)

	expense, err := s.GetExpenseById(trContext, expenseId)
	if err != nil {
		s.logger.Warn("expense not found for deletion",
			slog.String("expense_id", expenseId),
			slog.String("component", "service"),
		)
		return "", common.ErrExpenseNotFound
	}

	err = s.expensesRepository.DeleteExpenseById(trContext, expense.ExpenseId)
	if err != nil {
		s.logger.Error("error deleting expense",
			slog.String("expense_id", expenseId),
			slog.String("error", err.Error()),
			slog.String("component", "service"),
		)
		return "", err
	}

	s.logger.Info("expense deleted successfully",
		slog.String("expense_id", expenseId),
		slog.String("component", "service"),
	)

	return expense.ExpenseId, nil
}

// Function to get expenses by userId or category

func (s *ExpensesService) GetExpByUserIdorCat(ctx context.Context, key string, value string) ([]models.Expense, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCGetExpByUserIdorCat")
	childSpan.SetAttributes(attribute.String("component", "service"))
	defer childSpan.End()

	s.logger.Debug("getting expenses by user id or category",
		slog.String("key", key),
		slog.String("value", value),
		slog.String("component", "service"),
	)

	expenses, err := s.expensesRepository.GetExpByUserIdorCat(trContext, key, value)
	if err != nil {
		s.logger.Debug("expenses not found by query",
			slog.String("key", key),
			slog.String("value", value),
			slog.String("component", "service"),
		)
		return expenses, err
	}

	s.logger.Debug("expenses retrieved successfully",
		slog.String("key", key),
		slog.Int("count", len(expenses)),
		slog.String("component", "service"),
	)

	return expenses, nil
}

// Function to validate if user exist and is active

func (s *ExpensesService) validateUser(userId string) bool {
	// OTel instrumentation
	//tr := otel.Tracer(common.ServiceName)
	//trContext, childSpan := tr.Start(ctx, "SVCValidateUser")
	//childSpan.SetAttributes(attribute.String("component", "service"))
	//defer childSpan.End()

	userBaseURL := fmt.Sprintf("http://user/user/%s", userId)
	user := models.User{}

	// Validate if User exist and is not blocked
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

	s.logger.Debug("user validated successfully",
		slog.String("user_id", userId),
		slog.String("component", "service"),
	)

	return true
}
