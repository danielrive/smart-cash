package handler

import (
	"log/slog"
	"net/http"

	"smart-cash/bank-service/internal/common"
	"smart-cash/bank-service/internal/handler/dto"
	"smart-cash/bank-service/internal/service"
	"smart-cash/bank-service/models"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
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
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(c.Request.Context(), "HandlerHandlePayment")
	defer childSpan.End()

	// Get validated body from middleware
	validatedBody, exists := c.Get("validatedBody")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
		h.logger.Error("validatedBody not found in context")
		return
	}
	
	paymentDTO, ok := validatedBody.(dto.PayExpenseRequest)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		h.logger.Error("failed to cast validatedBody to PayExpenseRequest")
		return
	}
	
	transaction := models.TransactionRequest{
		TransactionId: paymentDTO.TransactionId,
		ExpenseId:     paymentDTO.ExpenseId,
		Date:          paymentDTO.Date,
		Amount:        paymentDTO.Amount,
		UserId:        paymentDTO.UserId,
		Status:        paymentDTO.Status,
	}
	
	// init payment
	transactionResult, err := h.bankService.ProcessPayment(trContext, transaction)
	if err != nil {
		h.logger.Error("error processing payment",
			"error", err.Error(),
		)
		c.JSON(http.StatusNotImplemented, gin.H{"error": common.ErrInternalError})
		return
	}
	
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
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(c.Request.Context(), "HandlerGetUser")
	defer childSpan.End()

	userId := c.Param("userId")

	user, err := h.bankService.GetUser(trContext, userId)

	if err != nil {
		h.logger.Error("error getting user",
			"error", err.Error(),
		)
		c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
		return
	}
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
