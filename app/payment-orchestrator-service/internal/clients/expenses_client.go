package clients

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"smart-cash/payment-orchestrator-service/models"
	"smart-cash/utils/logging"

	"github.com/go-resty/resty/v2"
)

type ExpensesClient struct {
	client  *resty.Client
	baseURL string
	logger  *slog.Logger
}

// loggerWithTrace returns a logger with trace context if available in the context
func (c *ExpensesClient) loggerWithTrace(ctx context.Context) *slog.Logger {
	return logging.LoggerWithTraceContext(ctx, c.logger)
}

func NewExpensesClient(baseURL string, logger *slog.Logger) *ExpensesClient {
	client := resty.New().
		SetTimeout(10 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)

	return &ExpensesClient{
		client:  client,
		baseURL: baseURL,
		logger:  logger,
	}
}

// UpdateExpenseStatus updates the status of an expense

func (c *ExpensesClient) UpdateExpenseStatus(ctx context.Context, expenseId string, status string) error {
	url := fmt.Sprintf("%s/expenses/%s/status", c.baseURL, expenseId)

	req := models.UpdateExpenseStatusRequest{
		Status: status,
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(req).
		Put(url)

	if err != nil {
		c.loggerWithTrace(ctx).Error("failed to update expense status",
			"error", err,
			"expense_id", expenseId,
			"status", status,
		)
		return fmt.Errorf("expenses service error: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		c.loggerWithTrace(ctx).Error("expenses service returned error",
			"status_code", resp.StatusCode(),
			"expense_id", expenseId,
			"status", status,
		)
		return fmt.Errorf("expenses service error: status %d", resp.StatusCode())
	}

	c.logger.Debug("expense status updated successfully",
		"expense_id", expenseId,
		"status", status,
	)

	return nil
}
