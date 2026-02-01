package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"smart-cash/payment-orchestrator-service/internal/clients"
	"smart-cash/payment-orchestrator-service/internal/common"
	"smart-cash/payment-orchestrator-service/models"
	"smart-cash/utils"
	"smart-cash/utils/logging"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type OrchestratorService struct {
	expensesClient *clients.ExpensesClient
	bankClient     *clients.BankClient
	paymentClient  *clients.PaymentClient
	logger         *slog.Logger
}

// loggerWithTrace returns a logger with trace context if available in the context
func (s *OrchestratorService) loggerWithTrace(ctx context.Context) *slog.Logger {
	return logging.LoggerWithTraceContext(ctx, s.logger)
}

func NewOrchestratorService(
	expensesClient *clients.ExpensesClient,
	bankClient *clients.BankClient,
	paymentClient *clients.PaymentClient,
	logger *slog.Logger,
) *OrchestratorService {
	return &OrchestratorService{
		expensesClient: expensesClient,
		bankClient:     bankClient,
		paymentClient:  paymentClient,
		logger:         logger,
	}
}

// ProcessPaymentEvent orchestrates the payment processing flow using Saga pattern
func (s *OrchestratorService) ProcessPaymentEvent(ctx context.Context, event models.PaymentEvent) error {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "OrchestratorProcessPaymentEvent", "service",
		attribute.String("payment.id", event.PaymentId),
		attribute.String("expense.id", event.ExpenseId),
		attribute.String("user.id", event.UserId),
		attribute.Float64("payment.amount", event.Amount),
		attribute.String("payment.status", event.Status),
	)
	defer endSpan()

	s.loggerWithTrace(ctx).Info("starting payment orchestration",
		"payment_id", event.PaymentId,
		"expense_id", event.ExpenseId,
		"user_id", event.UserId,
		"amount", event.Amount,
	)

	// Track compensation actions for rollback
	var compensationActions []func(context.Context) error

	// Step 1: Update payment status to "processing"
	utils.AddSpanEvent(ctx, "step 1: updating payment status to processing",
		attribute.String("payment.id", event.PaymentId),
	)
	if err := s.paymentClient.UpdatePaymentStatus(ctx, event.PaymentId, "processing"); err != nil {
		utils.RecordSpanError(ctx, err)
		s.loggerWithTrace(ctx).Error("failed to update payment status to processing",
			"error", err,
			"payment_id", event.PaymentId,
		)
		return fmt.Errorf("failed to update payment status: %w", err)
	}
	compensationActions = append(compensationActions, func(ctx context.Context) error {
		return s.paymentClient.UpdatePaymentStatus(ctx, event.PaymentId, "failed")
	})
	utils.AddSpanEvent(ctx, "step 1 completed: payment status set to processing",
		attribute.String("payment.id", event.PaymentId),
	)

	// Step 2: Update expense status to "paid"
	utils.AddSpanEvent(ctx, "step 2: updating expense status to paid",
		attribute.String("expense.id", event.ExpenseId),
	)
	if err := s.expensesClient.UpdateExpenseStatus(ctx, event.ExpenseId, "paid"); err != nil {
		utils.RecordSpanError(ctx, err)
		s.loggerWithTrace(ctx).Error("failed to update expense status to paid",
			"error", err,
			"expense_id", event.ExpenseId,
		)
		// Compensate: revert payment status
		s.compensate(ctx, compensationActions)
		return fmt.Errorf("failed to update expense status: %w", err)
	}
	compensationActions = append(compensationActions, func(ctx context.Context) error {
		return s.expensesClient.UpdateExpenseStatus(ctx, event.ExpenseId, "pending")
	})
	utils.AddSpanEvent(ctx, "step 2 completed: expense status set to paid",
		attribute.String("expense.id", event.ExpenseId),
	)

	// Step 3: Process bank transaction
	bankReq := models.ProcessTransactionRequest{
		UserId:        event.UserId,
		Amount:        event.Amount,
		TransactionId: event.PaymentId, // Use paymentId as transactionId
		ExpenseId:     event.ExpenseId,
	}

	utils.AddSpanEvent(ctx, "step 3: processing bank transaction",
		attribute.String("user.id", event.UserId),
		attribute.Float64("transaction.amount", event.Amount),
		attribute.String("transaction.id", event.PaymentId),
	)
	if err := s.bankClient.ProcessTransaction(ctx, bankReq); err != nil {
		utils.RecordSpanError(ctx, err)
		s.loggerWithTrace(ctx).Error("failed to process bank transaction",
			"error", err,
			"payment_id", event.PaymentId,
			"user_id", event.UserId,
			"amount", event.Amount,
		)
		// Compensate: revert expense and payment status
		// This will set expense → "pending" and payment → "failed"
		s.compensate(ctx, compensationActions)
		return fmt.Errorf("failed to process bank transaction: %w", err)
	}
	utils.AddSpanEvent(ctx, "step 3 completed: bank transaction processed",
		attribute.String("transaction.id", event.PaymentId),
	)

	// Step 4: Update payment status to "completed"
	utils.AddSpanEvent(ctx, "step 4: updating payment status to completed",
		attribute.String("payment.id", event.PaymentId),
	)
	if err := s.paymentClient.UpdatePaymentStatus(ctx, event.PaymentId, "completed"); err != nil {
		utils.RecordSpanError(ctx, err)
		s.loggerWithTrace(ctx).Error("failed to update payment status to completed",
			"error", err,
			"payment_id", event.PaymentId,
		)
		// Even if this fails, the transaction is already processed
		// We could implement a reconciliation job to fix this
		return fmt.Errorf("failed to update payment status to completed: %w", err)
	}
	utils.AddSpanEvent(ctx, "step 4 completed: payment status set to completed",
		attribute.String("payment.id", event.PaymentId),
	)

	utils.SetSpanStatus(ctx, codes.Ok, "payment orchestration completed successfully")

	s.loggerWithTrace(ctx).Info("payment orchestration completed successfully",
		"payment_id", event.PaymentId,
		"expense_id", event.ExpenseId,
		"user_id", event.UserId,
		"amount", event.Amount,
	)

	return nil
}

// compensate executes compensation actions in reverse order (rollback)
func (s *OrchestratorService) compensate(ctx context.Context, actions []func(context.Context) error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "OrchestratorCompensate", "service",
		attribute.Int("compensation.action_count", len(actions)),
	)
	defer endSpan()

	s.loggerWithTrace(ctx).Warn("executing compensation actions",
		"action_count", len(actions),
	)

	utils.AddSpanEvent(ctx, "starting compensation (rollback)",
		attribute.Int("compensation.action_count", len(actions)),
	)

	// Execute compensation actions in reverse order
	for i := len(actions) - 1; i >= 0; i-- {
		action := actions[i]
		utils.AddSpanEvent(ctx, "executing compensation action",
			attribute.Int("compensation.action_index", i),
		)
		if err := action(ctx); err != nil {
			utils.RecordSpanError(ctx, err)
			s.loggerWithTrace(ctx).Error("compensation action failed",
				"error", err,
				"action_index", i,
			)
			// Continue with other compensation actions even if one fails
		} else {
			utils.AddSpanEvent(ctx, "compensation action completed",
				attribute.Int("compensation.action_index", i),
			)
		}
	}

	utils.SetSpanStatus(ctx, codes.Ok, "compensation completed")
	utils.AddSpanEvent(ctx, "compensation (rollback) completed",
		attribute.Int("compensation.action_count", len(actions)),
	)
}

// ProcessPaymentEventWithRetry processes a payment event with retry logic
func (s *OrchestratorService) ProcessPaymentEventWithRetry(ctx context.Context, event models.PaymentEvent) error {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "OrchestratorProcessPaymentEventWithRetry", "service",
		attribute.String("payment.id", event.PaymentId),
		attribute.String("expense.id", event.ExpenseId),
		attribute.String("user.id", event.UserId),
	)
	defer endSpan()

	maxRetries := 3
	retryDelay := 1 * time.Second

	utils.AddSpanEvent(ctx, "starting retry logic",
		attribute.Int("retry.max_attempts", maxRetries),
		attribute.String("payment.id", event.PaymentId),
	)

	for attempt := 1; attempt <= maxRetries; attempt++ {
		utils.AddSpanEvent(ctx, "attempting payment orchestration",
			attribute.Int("retry.attempt", attempt),
			attribute.Int("retry.max_attempts", maxRetries),
		)

		err := s.ProcessPaymentEvent(ctx, event)
		if err == nil {
			utils.AddSpanEvent(ctx, "payment orchestration succeeded",
				attribute.Int("retry.attempt", attempt),
			)
			utils.SetSpanStatus(ctx, codes.Ok, "payment orchestration completed successfully")
			return nil
		}

		if attempt < maxRetries {
			utils.AddSpanEvent(ctx, "payment orchestration failed, will retry",
				attribute.Int("retry.attempt", attempt),
				attribute.Int("retry.max_attempts", maxRetries),
				attribute.String("error", err.Error()),
			)
			s.loggerWithTrace(ctx).Warn("payment orchestration failed, retrying",
				"payment_id", event.PaymentId,
				"attempt", attempt,
				"max_retries", maxRetries,
				"error", err,
			)
			time.Sleep(retryDelay * time.Duration(attempt))
		} else {
			utils.RecordSpanError(ctx, err)
			utils.AddSpanEvent(ctx, "payment orchestration failed after all retries",
				attribute.Int("retry.attempt", attempt),
				attribute.Int("retry.max_attempts", maxRetries),
			)
			s.loggerWithTrace(ctx).Error("payment orchestration failed after all retries",
				"payment_id", event.PaymentId,
				"attempts", maxRetries,
				"error", err,
			)
			return fmt.Errorf("payment orchestration failed after %d attempts: %w", maxRetries, err)
		}
	}

	return fmt.Errorf("unexpected error in retry logic")
}
