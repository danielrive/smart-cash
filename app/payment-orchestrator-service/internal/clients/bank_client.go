package clients

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"smart-cash/payment-orchestrator-service/models"
	"smart-cash/utils/logging"

	"github.com/go-resty/resty/v2"
)

type BankClient struct {
	client  *resty.Client
	baseURL string
	logger  *slog.Logger
}

// loggerWithTrace returns a logger with trace context if available in the context
func (c *BankClient) loggerWithTrace(ctx context.Context) *slog.Logger {
	return logging.LoggerWithTraceContext(ctx, c.logger)
}

func NewBankClient(baseURL string, logger *slog.Logger) *BankClient {
	client := resty.New().
		SetTimeout(10 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)

	return &BankClient{
		client:  client,
		baseURL: baseURL,
		logger:  logger,
	}
}

// ProcessTransaction processes a bank transaction
// Calls: POST /bank/pay
func (c *BankClient) ProcessTransaction(ctx context.Context, req models.ProcessTransactionRequest) error {
	url := fmt.Sprintf("%s/bank/pay", c.baseURL)

	// Map to the DTO format expected by bank service
	bankReq := map[string]interface{}{
		"transactionId": req.TransactionId,
		"expenseId":     req.ExpenseId,
		"userId":        req.UserId,
		"amount":        req.Amount,
		"status":        "pending",
		"date":          "", // Will be set by bank service if needed
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetBody(bankReq).
		Post(url)

	if err != nil {
		c.loggerWithTrace(ctx).Error("failed to process bank transaction",
			"error", err,
			"transaction_id", req.TransactionId,
			"user_id", req.UserId,
			"amount", req.Amount,
		)
		return fmt.Errorf("bank service error: %w", err)
	}

	if resp.StatusCode() != http.StatusCreated && resp.StatusCode() != http.StatusOK {
		body := resp.String()
		c.loggerWithTrace(ctx).Error("bank service returned error",
			"status_code", resp.StatusCode(),
			"response_body", body,
			"transaction_id", req.TransactionId,
		)
		return fmt.Errorf("bank service error: status %d", resp.StatusCode())
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp.Body(), &result); err == nil {
		if status, ok := result["status"].(string); ok {
			if status == "NotPaid" {
				return fmt.Errorf("bank transaction failed: insufficient funds")
			}
		}
	}

	c.logger.Debug("bank transaction processed successfully",
		"transaction_id", req.TransactionId,
		"user_id", req.UserId,
		"amount", req.Amount,
	)

	return nil
}
