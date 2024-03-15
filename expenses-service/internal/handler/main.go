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
	// create the expense
	if err := h.expensesService.CreateExpense(expense); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not created"})
		return
	}
	c.JSON(http.StatusOK, "ok")

}

// Handler to get expenses by tag

func (h *ExpensesHandler) GetExpensesByTag(c *gin.Context) {
	tag := c.Query("tag")
	userId := c.Query("userId")

	expenses, err := h.expensesService.GetExpensesByTag(tag, userId)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Expenses not found"})
		return
	}
	c.JSON(http.StatusOK, expenses)
}

func (h *ExpensesHandler) CalculateTotalPerCategory(c *gin.Context) {

	uri := c.Request.URL.Query()

	totalExpenses, err := h.expensesService.CalculateTotalPerCategory(uri["id"][0], uri["category"][0])

	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknow"})
		return
	}
	c.JSON(http.StatusOK, totalExpenses)
}
