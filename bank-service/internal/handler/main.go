package handler

import (
	"net/http"

	"bank-service/internal/models"
	"bank-service/internal/service"

	"github.com/gin-gonic/gin"
)

type bankHandler struct {
	bankService *service.BankService
}

func NewBankHandler(bankService *service.BankService) *bankHandler {
	return &bankHandler{bankService: bankService}
}

// Handler for creating new user

func (h *bankHandler) CreateTransaction(c *gin.Context) {
	transaction := models.Transaction{}
	// bind the JSON data to the user struct
	if err := c.ShouldBindJSON(&transaction); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// create the expense
	if err := h.bankService.CreateTransaction(transaction); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "transaction not created"})
		return
	}
	c.JSON(http.StatusOK, "ok")

}
