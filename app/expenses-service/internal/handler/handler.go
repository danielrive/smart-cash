package handler

import (
	"net/http"

	"smart-cash/expenses-service/internal/common"
	"smart-cash/expenses-service/internal/models"
	"smart-cash/expenses-service/internal/service"

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
		c.JSON(http.StatusNotImplemented, gin.H{"error": "expenses not created"})
		return
	}
	c.JSON(http.StatusCreated, "ok")

}

// Handler for Get method

func (h *ExpensesHandler) GetExpenses(c *gin.Context) {
	// validate the query in the url to see with what attribute filter
	uri := c.Request.URL.Query()

	// if userId is not present, then we need to return an error
	// if tag is present, then we need to get expenses by category, otherwise get all expenses by userId

	if _, isMapContainsKey := uri["userId"]; !isMapContainsKey {
		c.JSON(http.StatusBadRequest, gin.H{"error": "userId is required"})
		return
	} else if _, isMapContainsKey := uri["category"]; isMapContainsKey {
		expenses, err := h.expensesService.GetExpensesByCategory(uri["category"][0], uri["userId"][0])
		if err != nil {
			if err == common.ErrExpenseNotFound {
				c.JSON(http.StatusNotFound, gin.H{"Message": common.ErrExpenseNotFound})
				return
			} else if err == common.ErrWrongCredentials {
				c.JSON(http.StatusUnauthorized, gin.H{"error": common.ErrWrongCredentials})
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				return
			}
		}
		c.JSON(http.StatusOK, expenses)
		return
	} else {
		expenses, err := h.expensesService.GetExpensesByUserId(uri["userId"][0])
		if err != nil {
			if err == common.ErrExpenseNotFound {
				c.JSON(http.StatusNotFound, gin.H{"Message": common.ErrExpenseNotFound})
				return
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
				return
			}
		}
		c.JSON(http.StatusOK, expenses)
	}
}

/// Health check

func (h *ExpensesHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}

func (h *ExpensesHandler) ConnectToOtherSvc(c *gin.Context) {

	uri := c.Request.URL.Query()

	err := h.expensesService.ConnectOtherSVC(uri["svcName"][0], uri["port"][0])

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
		return
	}
	c.JSON(http.StatusOK, "ok")

}
