package handler

import (
	"log/slog"
	"net/http"

	"smart-cash/payment-service/internal/common"
	"smart-cash/payment-service/internal/handler/dto"
	"smart-cash/payment-service/internal/service"
	"smart-cash/payment-service/models"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
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
	// OTel Instrumentation
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(c.Request.Context(), "HandlerProcessPayment")
	defer childSpan.End()

	// Get validated body from middleware
	validatedBody, exists := c.Get("validatedBody")
	if !exists {
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
		h.logger.Error("validatedBody not found in context")
		return
	}
	
	// Type assert to our DTO
	paymentDTO, ok := validatedBody.(dto.ProcessPaymentRequest)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		h.logger.Error("failed to cast validatedBody to ProcessPaymentRequest")
		return
	}
	
	transaction := models.PaymentRequest{
		UserId:    paymentDTO.UserId,
		ExpenseId: paymentDTO.ExpenseId,
	}
	
	// init payment
	transactionResult, err := h.paymentService.ProcessPayment(trContext, transaction)
	if err != nil {
		h.logger.Error("error processing payment",
			"error", err.Error(),
			"level", "handler",
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

func (h *PaymentHandler) GetTransaction(c *gin.Context) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(c.Request.Context(), "HandlerGetTransaction")
	defer childSpan.End()

	transactionId := c.Param("transactionId")
	transaction, err := h.paymentService.GetTransaction(trContext, transactionId)

	if err != nil {
		if err == common.ErrTransactionNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": common.ErrTransactionNotFound})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError})
			return
		}

	}
	
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
