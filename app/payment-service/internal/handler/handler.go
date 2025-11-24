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
		slog.Float64("amount", paymentDTO.Amount),
		slog.String("component", "handler"),
	)

	// Convert DTO to model for service layer
	paymentRequest := models.PaymentRequest{
		UserId:    paymentDTO.UserId,
		ExpenseId: paymentDTO.ExpenseId,
		Amount:    paymentDTO.Amount,
	}

	// init payment
	paymentResult, err := h.paymentService.ProcessPayment(ctx, paymentRequest)
	if err != nil {
		h.logger.Error("error processing payment",
			slog.String("error", err.Error()),
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError.Error()})
		return
	}

	h.logger.Info("payment processed successfully",
		slog.String("payment_id", paymentResult.PaymentId),
		slog.String("expense_id", paymentResult.ExpenseId),
		slog.String("status", paymentResult.Status),
		slog.String("component", "handler"),
	)

	response := models.PaymentResponse{
		PaymentId: paymentResult.PaymentId,
		ExpenseId: paymentResult.ExpenseId,
		Date:      paymentResult.Date,
		Status:    paymentResult.Status,
	}

	c.JSON(http.StatusCreated, response)
}

func (h *PaymentHandler) GetPayment(c *gin.Context) {
	ctx, endSpan := utils.StartSpanWithComponent(c.Request.Context(), common.ServiceName, "HandlerGetPayment", "handler")
	defer endSpan()

	paymentId := c.Param("paymentId")

	h.logger.Info("getting payment",
		slog.String("payment_id", paymentId),
		slog.String("component", "handler"),
	)

	payment, err := h.paymentService.GetPayment(ctx, paymentId)
	if err != nil {
		if err == common.ErrPaymentNotFound {
			h.logger.Warn("payment not found",
				slog.String("payment_id", paymentId),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusNotFound, gin.H{"message": common.ErrPaymentNotFound})
			return
		} else {
			h.logger.Error("error getting payment",
				slog.String("payment_id", paymentId),
				slog.String("error", err.Error()),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError.Error()})
			return
		}
	}

	h.logger.Info("payment retrieved successfully",
		slog.String("payment_id", paymentId),
		slog.String("status", payment.Status),
		slog.String("component", "handler"),
	)

	response := models.PaymentResponse{
		PaymentId: payment.PaymentId,
		ExpenseId: payment.ExpenseId,
		Date:      payment.Date,
		Status:    payment.Status,
	}

	c.JSON(http.StatusOK, response)
}

func (h *PaymentHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}
