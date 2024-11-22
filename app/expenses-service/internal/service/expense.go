package service

import (
	"bytes"
	"context"
	"encoding/json"
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

// Define service interface

type ExpensesService struct {
	expensesRepository *repositories.DynamoDBExpensesRepository
	logger             *slog.Logger
}

// Create a new expenses service
func NewExpensesService(expensesRepository *repositories.DynamoDBExpensesRepository, logger *slog.Logger) *ExpensesService {
	return &ExpensesService{
		expensesRepository: expensesRepository,
		logger:             logger,
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

// Function to process expenses
func (s *ExpensesService) PayExpenses(ctx context.Context, expensesId models.ExpensesPay) (models.Expense, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCPayExpenses")
	childSpan.SetAttributes(attribute.String("component", "service"))
	defer childSpan.End()

	baseURL := "http://bank/bank/pay"
	// get the expense from DB
	expense, err := s.GetExpenseById(trContext, expensesId.ExpenseId)
	if err != nil {
		return models.Expense{}, common.ErrInternalError
	}

	s.logger.Info("preparing request to pay",
		"user", expense.UserId,
	)

	// send the expenses to payment services sync proccess
	// create payment request per expenses
	paymentRequest := models.PaymentRequest{
		ExpenseId: expense.ExpenseId,
		Date:      time.Now().UTC().Format("2006-01-02"),
		UserId:    expense.UserId,
		Amount:    expense.Amount,
		Status:    expense.Status,
	}
	jsonData, err := json.Marshal(paymentRequest)

	if err != nil {
		s.logger.Error("Error marshalling data to JSON",
			"error", err.Error(),
		)
		return models.Expense{}, common.ErrInternalError
	}
	// Prepare the request for bank service
	req, err := http.NewRequest(http.MethodPost, baseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		s.logger.Error("error creating the http request",
			"error", err.Error(),
			"url", baseURL,
		)
		return models.Expense{}, common.ErrInternalError
	}
	// set headers
	req.Header.Set("Content-Type", "application/json")
	// send the request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		s.logger.Error("error sending http request",
			"error", err.Error(),
			"url", baseURL,
		)
		return models.Expense{}, common.ErrInternalError
	}
	s.logger.Info("request processed",
		"http_status", resp.StatusCode,
	)
	// validate response code
	if resp.StatusCode != http.StatusCreated {
		s.logger.Info("expense not paid",
			"http_status", resp.StatusCode,
			"expenseId", expense.ExpenseId,
		)
		return models.Expense{}, common.ErrExpenseNotPaid
	} else {
		resBody, err := io.ReadAll(resp.Body)
		if err != nil {
			s.logger.Error("error could not read response body",
				"error", err.Error(),
			)
			//// HOW to manage this kind of errors when the request already was procesed by another service but
			// for some situation like server error faild in the service that called
			return models.Expense{}, common.ErrInternalError
		}
		err = json.Unmarshal(resBody, &paymentRequest)
		if err != nil {
			s.logger.Error("error could not parse response body",
				"error", err.Error(),
			)
			return models.Expense{}, common.ErrInternalError
		}
		expense.Status = paymentRequest.Status
	}
	// Process response
	s.logger.Info("updating expense status to paid",
		"expensesId", expense.ExpenseId,
	)
	s.expensesRepository.UpdateExpenseStatus(trContext, expense)

	return expense, nil
}
