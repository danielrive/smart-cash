package handler

import (
	"log"
	"net/http"

	"smart-cash/bank-service/internal/models"
	"smart-cash/bank-service/internal/service"

	"github.com/gin-gonic/gin"
)

type BankHandler struct {
	bankService *service.BankService
}

func NewBankHandler(bankService *service.BankService) *BankHandler {
	return &BankHandler{bankService: bankService}
}

// Handler for creating new user

func (h *BankHandler) HandlePayment(c *gin.Context) {
	transaction := models.PaymentRequest{}
	// bind the JSON data to the user struct
	if err := c.ShouldBindJSON(&transaction); err != nil {
		log.Printf("error binding body to json %v:", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// init payment
	response, err := h.bankService.ProcessPayment(transaction)
	if err != nil {
		log.Printf("error processing payment %v:", err)
		c.JSON(http.StatusNotImplemented, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, response)
}

func (h *BankHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}
