package handler

import (
	"log/slog"
	"net/http"

	"smart-cash/bank-service/internal/common"
	"smart-cash/bank-service/internal/models"
	"smart-cash/bank-service/internal/service"

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
	response, err := h.bankService.ProcessPayment(transaction)
	if err != nil {
		h.logger.Error("error processing payment",
			"error", err.Error(),
		)
		c.JSON(http.StatusNotImplemented, gin.H{"error": common.ErrInternalError})
		return
	}
	c.JSON(http.StatusCreated, response)
}

func (h *BankHandler) ValidateUser(c *gin.Context) {
	userId := c.Param("userId")

	user, err := h.bankService.GetUser(userId)

	if err != nil {
		h.logger.Error("error getting user",
			"error", err.Error(),
		)
		c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *BankHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}
