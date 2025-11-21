package handler

import (
	"log/slog"
	"net/http"

	"smart-cash/payment-service/internal/common"
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

	paymentRequest := models.PaymentRequest{}
	// bind the JSON data to the user struct
	if err := c.ShouldBindJSON(&paymentRequest); err != nil {
		h.logger.Error("error binding json",
			"error", err.Error(),
			"level", "handler",
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	h.logger.Info("processing payment",
		slog.String("user_id", paymentRequest.UserId),
		slog.String("expense_id", paymentRequest.ExpenseId),
		slog.String("component", "handler"),
	)

	// init payment
	transactionResult, err := h.paymentService.ProcessPayment(ctx, paymentRequest)
	if err != nil {
		h.logger.Error("error processing payment",
			"error", err.Error(),
			"level", "handler",
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError})
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
			c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError})
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
