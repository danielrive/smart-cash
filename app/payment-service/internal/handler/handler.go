package handler

import (
	"log/slog"
	"net/http"

	"smart-cash/payment-service/internal/common"
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
	trContext, childSpan := tr.Start(c.Request.Context(), "HandlerGetTransaction")
	defer childSpan.End()

	transaction := models.PaymentRequest{}
	// bind the JSON data to the user struct
	if err := c.ShouldBindJSON(&transaction); err != nil {
		h.logger.Error("error binding json",
			"error", err.Error(),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	// init payment
	response, err := h.paymentService.ProcessPayment(trContext, transaction)
	if err != nil {
		h.logger.Error("error processing payment",
			"error", err.Error(),
		)
		c.JSON(http.StatusNotImplemented, gin.H{"error": common.ErrInternalError})
		return
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
	c.JSON(http.StatusOK, transaction)
}

func (h *PaymentHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}
