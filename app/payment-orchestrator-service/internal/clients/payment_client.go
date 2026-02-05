package clients

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"smart-cash/payment-orchestrator-service/models"
	"smart-cash/utils/logging"
	"smart-cash/utils/middleware"

	"github.com/go-resty/resty/v2"
)

type PaymentClient struct {
	client      *resty.Client
	baseURL     string
	logger      *slog.Logger
	jwtSecret   []byte
	serviceName string
}

// loggerWithTrace returns a logger with trace context if available in the context
func (c *PaymentClient) loggerWithTrace(ctx context.Context) *slog.Logger {
	return logging.LoggerWithTraceContext(ctx, c.logger)
}

func NewPaymentClient(baseURL string, logger *slog.Logger, jwtSecret []byte, serviceName string) *PaymentClient {
	client := resty.New().
		SetTimeout(10 * time.Second).
		SetRetryCount(3).
		SetRetryWaitTime(1 * time.Second).
		SetRetryMaxWaitTime(5 * time.Second)

	return &PaymentClient{
		client:      client,
		baseURL:     baseURL,
		logger:      logger,
		jwtSecret:   jwtSecret,
		serviceName: serviceName,
	}
}

// UpdatePaymentStatus updates the status of a payment
// NOTE: This endpoint may need to be added to payment-service
// Expected endpoint: PUT /payment/:paymentId/status
func (c *PaymentClient) UpdatePaymentStatus(ctx context.Context, paymentId string, status string) error {
	url := fmt.Sprintf("%s/payment/%s/status", c.baseURL, paymentId)

	req := models.UpdatePaymentStatusRequest{
		Status: status,
	}

	// Generate service JWT token for authentication
	token, err := middleware.GenerateServiceJWT(c.serviceName, c.jwtSecret)
	if err != nil {
		c.loggerWithTrace(ctx).Error("failed to generate service JWT token",
			"error", err,
			"payment_id", paymentId,
		)
		return fmt.Errorf("failed to generate service JWT token: %w", err)
	}

	resp, err := c.client.R().
		SetContext(ctx).
		SetHeader("Content-Type", "application/json").
		SetAuthToken(token).
		SetBody(req).
		Put(url)

	if err != nil {
		c.loggerWithTrace(ctx).Error("failed to update payment status",
			"error", err,
			"payment_id", paymentId,
			"status", status,
		)
		return fmt.Errorf("payment service error: %w", err)
	}

	if resp.StatusCode() != http.StatusOK {
		c.loggerWithTrace(ctx).Error("payment service returned error",
			"status_code", resp.StatusCode(),
			"payment_id", paymentId,
			"status", status,
		)
		return fmt.Errorf("payment service error: status %d", resp.StatusCode())
	}

	c.loggerWithTrace(ctx).Debug("payment status updated successfully",
		"payment_id", paymentId,
		"status", status,
	)

	return nil
}
