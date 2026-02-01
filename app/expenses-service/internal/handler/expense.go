package handler

import (
	"log/slog"
	"net/http"
	"time"

	"smart-cash/expenses-service/internal/common"
	"smart-cash/expenses-service/internal/handler/dto"
	"smart-cash/expenses-service/internal/service"
	"smart-cash/expenses-service/models"
	"smart-cash/utils"

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
	ctx, endSpan := utils.StartSpanWithComponent(c.Request.Context(), common.ServiceName, "HandlerDeleteExpense", "handler")
	defer endSpan()

	expenseId := c.Param("expenseId")

	h.logger.Info("deleting expense",
		slog.String("expense_id", expenseId),
		slog.String("component", "handler"),
	)

	expense, err := h.expensesService.DeleteExpense(ctx, expenseId)
	if err != nil {
		if err == common.ErrExpenseNotFound {
			h.logger.Warn("expense not found for deletion",
				slog.String("expense_id", expenseId),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": common.ErrExpenseNotFound})
		} else {
			h.logger.Error("error deleting expense",
				slog.String("expense_id", expenseId),
				slog.String("error", err.Error()),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError})
		}
		return
	}

	h.logger.Info("expense deleted successfully",
		slog.String("expense_id", expense),
		slog.String("component", "handler"),
	)

	c.JSON(http.StatusOK, gin.H{"expenseId": expense})
}

// Handler for creating new user

func (h *ExpensesHandler) CreateExpense(c *gin.Context) {
	ctx, endSpan := utils.StartSpanWithComponent(c.Request.Context(), common.ServiceName, "HandlerCreateExpense", "handler")
	defer endSpan()

	// Get userId from auth middleware (set by AuthMiddleware)
	userId, exists := c.Get("userId")
	if !exists {
		h.logger.Error("userId not found in context",
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user not authenticated"})
		return
	}

	// Get validated body from validation middleware
	validatedBody, exists := c.Get("validatedBody")
	if !exists {
		h.logger.Error("validatedBody not found in context",
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
		return
	}

	// Type assert to our DTO
	expenseRequest, ok := validatedBody.(dto.CreateExpenseRequest)
	if !ok {
		h.logger.Error("failed to cast validatedBody to CreateExpenseRequest",
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	// Set userId from auth context
	expenseRequest.UserId = userId.(string)

	h.logger.Info("creating new expense",
		slog.String("user_id", expenseRequest.UserId),
		slog.String("name", expenseRequest.Name),
		slog.Float64("amount", expenseRequest.Amount),
		slog.String("category", expenseRequest.Category),
		slog.String("component", "handler"),
	)

	var expenseDate time.Time
	if expenseRequest.Date != "" {
		parsedDate, err := time.Parse("2006-01-02", expenseRequest.Date)
		if err != nil {
			h.logger.Error("failed to parse date",
				slog.String("error", err.Error()),
				slog.String("date", expenseRequest.Date),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid date format"})
			return
		}
		expenseDate = parsedDate
	} else {
		expenseDate = time.Now().UTC()
	}

	expense := models.Expense{
		UserId:      expenseRequest.UserId,
		Name:        expenseRequest.Name,
		Amount:      expenseRequest.Amount,
		Description: expenseRequest.Description,
		Category:    expenseRequest.Category,
		Date:        expenseDate,
		Tags:        expenseRequest.Tags,
	}

	// create the expense
	response, err := h.expensesService.CreateExpense(ctx, expense)
	if err != nil {
		h.logger.Error("error creating expense",
			slog.String("user_id", expenseRequest.UserId),
			slog.String("error", err.Error()),
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError})
		return
	}

	h.logger.Info("expense created successfully",
		slog.String("expense_id", response.ExpenseId),
		slog.String("user_id", expenseRequest.UserId),
		slog.String("component", "handler"),
	)

	c.Header("Location", "/expense/"+response.ExpenseId)
	c.JSON(http.StatusCreated, response)

}

// Handler for Get expense by expenseID

func (h *ExpensesHandler) GetExpensesById(c *gin.Context) {
	ctx, endSpan := utils.StartSpanWithComponent(c.Request.Context(), common.ServiceName, "HandlerGetExpensesById", "handler")
	defer endSpan()

	expenseId := c.Param("expenseId")

	h.logger.Info("getting expense by id",
		slog.String("expense_id", expenseId),
		slog.String("component", "handler"),
	)

	expenses, err := h.expensesService.GetExpenseById(ctx, expenseId)
	if err != nil {
		if err == common.ErrExpenseNotFound {
			h.logger.Warn("expense not found",
				slog.String("expense_id", expenseId),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusNotFound, gin.H{"message": common.ErrExpenseNotFound})
			return
		} else {
			h.logger.Error("error getting expense",
				slog.String("expense_id", expenseId),
				slog.String("error", err.Error()),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError})
			return
		}
	}

	h.logger.Info("expense retrieved successfully",
		slog.String("expense_id", expenseId),
		slog.String("component", "handler"),
	)

	c.JSON(http.StatusOK, expenses)
}

func (h *ExpensesHandler) GetExpensesByQuery(c *gin.Context) {
	ctx, endSpan := utils.StartSpanWithComponent(c.Request.Context(), common.ServiceName, "HandlerGetExpensesByQuery", "handler")
	defer endSpan()

	// validate the query in the url to see with what attribute filter
	query := c.Request.URL.Query()
	var key, value string
	if userId, ok := query["userId"]; ok {
		key, value = "userId", userId[0]
	} else if category, ok := query["category"]; ok {
		key, value = "category", category[0]
	} else {
		h.logger.Warn("invalid query parameters",
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Bad request"})
		return
	}

	h.logger.Info("getting expenses by query",
		slog.String("key", key),
		slog.String("value", value),
		slog.String("component", "handler"),
	)

	expenses, err := h.expensesService.GetExpByUserIdorCat(ctx, key, value)
	if err != nil {
		if err == common.ErrExpenseNotFound {
			h.logger.Warn("expenses not found by query",
				slog.String("key", key),
				slog.String("value", value),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusNotFound, gin.H{"Message": common.ErrExpenseNotFound})
			return
		} else if err == common.ErrWrongCredentials {
			h.logger.Warn("wrong credentials for expense query",
				slog.String("key", key),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusUnauthorized, gin.H{"error": common.ErrWrongCredentials})
			return
		} else {
			h.logger.Error("error getting expenses by query",
				slog.String("key", key),
				slog.String("value", value),
				slog.String("error", err.Error()),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
			return
		}
	}

	h.logger.Info("expenses retrieved by query successfully",
		slog.String("key", key),
		slog.String("value", value),
		slog.Int("count", len(expenses)),
		slog.String("component", "handler"),
	)

	c.JSON(http.StatusOK, expenses)
}

// Handler for updating expense status
func (h *ExpensesHandler) UpdateExpenseStatus(c *gin.Context) {
	ctx, endSpan := utils.StartSpanWithComponent(c.Request.Context(), common.ServiceName, "HandlerUpdateExpenseStatus", "handler")
	defer endSpan()

	expenseId := c.Param("expenseId")

	// Get validated body from validation middleware
	validatedBody, exists := c.Get("validatedBody")
	if !exists {
		h.logger.Error("validatedBody not found in context",
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": "validation failed"})
		return
	}

	// Type assert to our DTO
	statusRequest, ok := validatedBody.(dto.UpdateExpenseStatusRequest)
	if !ok {
		h.logger.Error("failed to cast validatedBody to UpdateExpenseStatusRequest",
			slog.String("component", "handler"),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal error"})
		return
	}

	h.logger.Info("updating expense status",
		slog.String("expense_id", expenseId),
		slog.String("status", statusRequest.Status),
		slog.String("component", "handler"),
	)

	response, err := h.expensesService.UpdateExpenseStatus(ctx, expenseId, statusRequest.Status)
	if err != nil {
		if err == common.ErrExpenseNotFound {
			h.logger.Warn("expense not found for status update",
				slog.String("expense_id", expenseId),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusNotFound, gin.H{"error": common.ErrExpenseNotFound.Error()})
		} else {
			h.logger.Error("error updating expense status",
				slog.String("expense_id", expenseId),
				slog.String("error", err.Error()),
				slog.String("component", "handler"),
			)
			c.JSON(http.StatusInternalServerError, gin.H{"error": common.ErrInternalError.Error()})
		}
		return
	}

	h.logger.Info("expense status updated successfully",
		slog.String("expense_id", expenseId),
		slog.String("status", statusRequest.Status),
		slog.String("component", "handler"),
	)

	c.JSON(http.StatusOK, response)
}

/// Health check

func (h *ExpensesHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, "ok")
}
