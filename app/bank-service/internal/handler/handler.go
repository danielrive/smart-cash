package handler

import (
	"log/slog"
	"net/http"

	"smart-cash/bank-service/internal/common"
	"smart-cash/bank-service/internal/handler/dto"
	"smart-cash/bank-service/internal/service"
	"smart-cash/bank-service/models"
	"smart-cash/utils"

	"github.com/gin-gonic/gin"
)

type BankHandler struct {
	bankService *service.BankService
	logger      *slog.Logger
}

func NewBankHandler(bankService *service.BankService, logger *slog.Logger) *BankHandler {
	return &BankHandler{
		bankService: bankService,
		logger:      logger,
	}
}

// Handler for creating new user

func (h *BankHandler) HandlePayment(c *gin.Context) {
	ctx, endSpan := utils.StartSpanWithComponent(c.Request.Context(), common.ServiceName, "HandlerHandlePayment", "handler")
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

	paymentDTO, ok := validatedBody.(dto.PayExpenseRequest)
	if !ok {
		h.logger.Error("failed to cast validatedBody to PayExpenseRequest",
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	h.logger.Info("handling bank payment",
		slog.String("transaction_id", paymentDTO.TransactionId),
		slog.String("expense_id", paymentDTO.ExpenseId),
		slog.String("user_id", paymentDTO.UserId),
		slog.Float64("amount", paymentDTO.Amount),
		slog.String("component", "handler"),
	)

	transaction := models.TransactionRequest{
		TransactionId: paymentDTO.TransactionId,
		ExpenseId:     paymentDTO.ExpenseId,
		Date:          paymentDTO.Date,
		Amount:        paymentDTO.Amount,
		UserId:        paymentDTO.UserId,
		Status:        paymentDTO.Status,
	}

	// init payment
	transactionResult, err := h.bankService.ProcessPayment(ctx, transaction)
	if err != nil {
		h.logger.Error("error processing bank payment",
			slog.String("transaction_id", paymentDTO.TransactionId),
			slog.String("error", err.Error()),
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError})
		return
	}

	h.logger.Info("bank payment processed successfully",
		slog.String("transaction_id", transactionResult.TransactionId),
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

func (h *BankHandler) GetUser(c *gin.Context) {
	ctx, endSpan := utils.StartSpanWithComponent(c.Request.Context(), common.ServiceName, "HandlerGetUser", "handler")
	defer endSpan()

	userId := c.Param("userId")

	h.logger.Info("getting bank user",
		slog.String("user_id", userId),
		slog.String("component", "handler"),
	)

	user, err := h.bankService.GetUser(ctx, userId)
	if err != nil {
		h.logger.Error("error getting bank user",
			slog.String("user_id", userId),
			slog.String("error", err.Error()),
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	h.logger.Info("bank user retrieved successfully",
		slog.String("user_id", userId),
		slog.String("currency", user.Currency),
		slog.String("component", "handler"),
	)

	response := models.BankUserResponse{
		UserId:   user.UserId,
		Currency: user.Currency,
		Savings:  user.Savings,
		Blocked:  user.Blocked,
	}

	c.JSON(http.StatusOK, response)
}

func (h *BankHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}
