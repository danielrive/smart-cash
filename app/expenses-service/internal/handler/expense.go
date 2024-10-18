package handler

import (
	"log/slog"
	"net/http"

	"smart-cash/expenses-service/internal/common"
	"smart-cash/expenses-service/internal/models"
	"smart-cash/expenses-service/internal/service"

	"github.com/gin-gonic/gin"
)

type ExpensesHandler struct {
	expensesService *service.ExpensesService
	logger          *slog.Logger
}

func NewExpensesHandler(expensesService *service.ExpensesService, logger *slog.Logger) *ExpensesHandler {
	return &ExpensesHandler{
		expensesService: expensesService,
		logger:          logger,
	}
}

// Handler for delete expense

func (h *ExpensesHandler) DeleteExpense(c *gin.Context) {
	expenseId := c.Param("expenseId")
	expense, err := h.expensesService.DeleteExpense(expenseId)

	if err != nil {
		if err == common.ErrExpenseNotFound {

			c.JSON(http.StatusNotImplemented, gin.H{"error": common.ErrExpenseNotFound})
		} else {
			c.JSON(http.StatusNotImplemented, gin.H{"error": common.ErrInternalError})
		}
	}
	c.JSON(http.StatusOK, gin.H{"expenseId": expense})
}

// Handler for creating new user

func (h *ExpensesHandler) CreateExpense(c *gin.Context) {
	expense := models.Expense{}
	// bind the JSON data to the user struct
	if err := c.ShouldBindJSON(&expense); err != nil {
		h.logger.Error("error binding json",
			"error", err.Error(),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	// create the expense
	response, err := h.expensesService.CreateExpense(expense)
	if err != nil {
		h.logger.Error("error processing expense",
			"error", err.Error(),
		)
		c.JSON(http.StatusNotImplemented, gin.H{"error": common.ErrInternalError})
		return
	}
	c.Header("Location", "/expense/"+response.ExpenseId)
	c.JSON(http.StatusCreated, response)

}

// Handler to pay expenses

func (h *ExpensesHandler) PayExpenses(c *gin.Context) {
	expenses := models.ExpensesPay{}
	if err := c.ShouldBindJSON(&expenses); err != nil {
		h.logger.Error("error binding json",
			"error", err.Error(),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
		return
	}
	response, err := h.expensesService.PayExpenses(expenses)
	if err != nil {
		h.logger.Error("error processing expense",
			"error", err.Error(),
		)
		c.JSON(http.StatusNotImplemented, gin.H{"error": common.ErrInternalError})
		return
	}
	c.JSON(http.StatusCreated, response)
}

// Handler for Get expense by expenseID

func (h *ExpensesHandler) GetExpensesById(c *gin.Context) {
	expenseId := c.Param("expenseId")

	expenses, err := h.expensesService.GetExpenseById(expenseId)
	if err != nil {
		if err == common.ErrExpenseNotFound {
			c.JSON(http.StatusNotFound, gin.H{"message": common.ErrExpenseNotFound})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError})
			return
		}
	}
	c.JSON(http.StatusOK, expenses)
}

func (h *ExpensesHandler) GetExpensesByQuery(c *gin.Context) {
	// validate the query in the url to see with what attribute filter
	query := c.Request.URL.Query()
	var key, value string
	if userId, ok := query["userId"]; ok {
		key, value = "userId", userId[0]
	} else if category, ok := query["category"]; ok {
		key, value = "category", category[0]
	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}
	expenses, err := h.expensesService.GetExpByUserIdorCat(key, value)
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
}

/// Health check

func (h *ExpensesHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}
