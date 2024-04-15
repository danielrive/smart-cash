package handler

import (
	"net/http"

	"smart-cash/bank-service/internal/common"
	"smart-cash/bank-service/internal/models"
	"smart-cash/bank-service/internal/service"

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

// get transactions by Id

func (h *bankHandler) GetTransactions(c *gin.Context) {
	uri := c.Request.URL.Query()

	// if orderId is not present, then we need to return an error
	// if tag is present, then we need to get payment by category, otherwise get all payment by orderId

	if _, isMapContainsKey := uri["id"]; !isMapContainsKey {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id is required"})
		return
	} else if _, isMapContainsKey := uri["id"]; isMapContainsKey {
		order, err := h.bankService.GetTransactions(uri["id"][0])
		if err != nil {
			if err == common.ErrTransactionNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": common.ErrTransactionNotFound})
				return
			} else if err == common.ErrWrongCredentials {
				c.JSON(http.StatusUnauthorized, gin.H{"error": common.ErrWrongCredentials})
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				return
			}
		}
		c.JSON(http.StatusOK, order)
		return
	}
}
