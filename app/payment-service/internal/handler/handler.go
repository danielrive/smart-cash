package handler

import (
	"log/slog"
	"net/http"

	"smart-cash/payment-service/internal/common"
	"smart-cash/payment-service/internal/handler/dto"
	"smart-cash/payment-service/internal/service"
	"smart-cash/payment-service/models"
	"smart-cash/utils"

	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *service.PaymentService
	logger         *slog.Logger
}

func NewPaymentHandler(paymentService *service.PaymentService, logger *slog.Logger) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		logger:         logger,
	}
}

// Handler for creating new user

func (h *PaymentHandler) ProcessPayment(c *gin.Context) {
	ctx, endSpan := utils.StartSpanWithComponent(c.Request.Context(), common.ServiceName, "HandlerProcessPayment", "handler")
	defer endSpan()

	// Get validated body from middleware
	validatedBody, exists := c.Get("validatedBody")
	if !exists {
		h.logger.Error("validatedBody not found in context",
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
		return
	}

	paymentDTO, ok := validatedBody.(dto.ProcessPaymentRequest)
	if !ok {
		h.logger.Error("failed to cast validatedBody to ProcessPaymentRequest",
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError.Error()})
		return
	}

	h.logger.Info("processing payment",
		slog.String("user_id", paymentDTO.UserId),
		slog.String("expense_id", paymentDTO.ExpenseId),
		slog.String("component", "handler"),
	)

	// Convert DTO to model for service layer
	paymentRequest := models.PaymentRequest{
		UserId:    paymentDTO.UserId,
		ExpenseId: paymentDTO.ExpenseId,
	}

	// init payment
	transactionResult, err := h.paymentService.ProcessPayment(ctx, paymentRequest)
	if err != nil {
		h.logger.Error("error processing payment",
			"error", err.Error(),
			"level", "handler",
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError.Error()})
		return
	}

	h.logger.Info("payment processed successfully",
		slog.String("transaction_id", transactionResult.TransactionId),
		slog.String("expense_id", transactionResult.ExpenseId),
		slog.String("status", transactionResult.Status),
		slog.String("component", "handler"),
	)

	response := models.TransactionResponse{
		TransactionId: transactionResult.TransactionId,
		ExpenseId:     transactionResult.ExpenseId,
		Date:          transactionResult.Date,
		Amount:        transactionResult.Amount,
		Status:        transactionResult.Status,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *PaymentHandler) GetTransaction(c *gin.Context) {
	ctx, endSpan := utils.StartSpanWithComponent(c.Request.Context(), common.ServiceName, "HandlerGetTransaction", "handler")
	defer endSpan()

	transactionId := c.Param("transactionId")

	h.logger.Info("getting transaction",
		slog.String("transaction_id", transactionId),
		slog.String("component", "handler"),
	)

	transaction, err := h.paymentService.GetTransaction(ctx, transactionId)
	if err != nil {
		if err == common.ErrTransactionNotFound {
			h.logger.Warn("transaction not found",
				slog.String("transaction_id", transactionId),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusNotFound, gin.H{"message": common.ErrTransactionNotFound})
			return
		} else {
			h.logger.Error("error getting transaction",
				slog.String("transaction_id", transactionId),
				slog.String("error", err.Error()),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError.Error()})
			return
		}
	}

	h.logger.Info("transaction retrieved successfully",
		slog.String("transaction_id", transactionId),
		slog.String("status", transaction.Status),
		slog.String("component", "handler"),
	)

	response := models.TransactionResponse{
		TransactionId: transaction.TransactionId,
		ExpenseId:     transaction.ExpenseId,
		Date:          transaction.Date,
		Amount:        transaction.Amount,
		Status:        transaction.Status,
	}

	c.JSON(http.StatusOK, response)
}

func (h *PaymentHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}
