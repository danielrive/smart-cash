package handler

import (
	"net/http"

	"expenses-service/internal/models"
	"expenses-service/internal/service"

	"github.com/gin-gonic/gin"
)

type ExpensesHandler struct {
	expensesService *service.ExpensesService
}

func NewExpensesHandler(expensesService *service.ExpensesService) *ExpensesHandler {
	return &ExpensesHandler{expensesService: expensesService}
}

// Handler for creating new user

func (h *ExpensesHandler) CreateExpense(c *gin.Context) {
	expense := models.Expense{}
	// bind the JSON data to the user struct
	if err := c.ShouldBindJSON(&expense); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// create the user
	if err := h.expensesService.CreateExpense(expense); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not created"})
		return
	}
	c.JSON(http.StatusOK, "ok")
}
