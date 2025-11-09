package service

import (
	"context"
	"log/slog"
	"smart-cash/expenses-service/internal/common"
	"smart-cash/expenses-service/internal/repositories"
	"smart-cash/expenses-service/models"
	"time"

	"go.opentelemetry.io/otel"
)

// ExpenseAnalytics represents aggregated expense data
type ExpenseAnalytics struct {
	TotalAmount     float64                 `json:"totalAmount"`
	CategoryTotals  map[string]float64      `json:"categoryTotals"`
	MonthlyTotals   map[string]float64      `json:"monthlyTotals"`
	TopExpenses     []models.ExpensesReturn `json:"topExpenses"`
	AverageExpense  float64                 `json:"averageExpense"`
}

// GetExpenseAnalytics calculates expense analytics for a given time period
func (s *ExpenseService) GetExpenseAnalytics(ctx context.Context, userID string, period string) (*ExpenseAnalytics, error) {
	tr := otel.Tracer(common.ServiceName)
	ctx, span := tr.Start(ctx, "ServiceGetExpenseAnalytics")
	defer span.End()

	// Calculate date range based on period
	startDate, endDate := calculateDateRange(period)

	// Get expenses for the period
	expenses, err := s.repo.GetExpensesByDateRange(ctx, userID, repositories.ExpenseFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	})
	if err != nil {
		s.logger.Error("error getting expenses for analytics",
			slog.String("error", err.Error()),
			slog.String("userId", userID),
			slog.String("period", period),
		)
		return nil, err
	}

	// Initialize analytics
	analytics := &ExpenseAnalytics{
		CategoryTotals: make(map[string]float64),
		MonthlyTotals:  make(map[string]float64),
	}

	// Process expenses
	var total float64
	for _, expense := range expenses {
		// Update total
		total += expense.Amount

		// Update category totals
		analytics.CategoryTotals[expense.Category] += expense.Amount

		// Update monthly totals
		monthKey := expense.Date.Format("2006-01")
		analytics.MonthlyTotals[monthKey] += expense.Amount
	}

	// Calculate average
	if len(expenses) > 0 {
		analytics.AverageExpense = total / float64(len(expenses))
	}

	analytics.TotalAmount = total

	// Get top expenses (top 5)
	topExpenses, err := s.getTopExpenses(ctx, userID, startDate, endDate, 5)
	if err != nil {
		s.logger.Error("error getting top expenses",
			slog.String("error", err.Error()),
			slog.String("userId", userID),
		)
		// Don't return error, continue with partial data
	} else {
		analytics.TopExpenses = topExpenses
	}

	return analytics, nil
}

// calculateDateRange returns start and end dates based on period
func calculateDateRange(period string) (string, string) {
	now := time.Now()
	endDate := now.Format("2006-01-02")

	switch period {
	case "week":
		startDate := now.AddDate(0, 0, -7).Format("2006-01-02")
		return startDate, endDate
	case "month":
		startDate := now.AddDate(0, -1, 0).Format("2006-01-02")
		return startDate, endDate
	case "year":
		startDate := now.AddDate(-1, 0, 0).Format("2006-01-02")
		return startDate, endDate
	default: // default to month
		startDate := now.AddDate(0, -1, 0).Format("2006-01-02")
		return startDate, endDate
	}
}

// getTopExpenses returns the top N expenses by amount
func (s *ExpenseService) getTopExpenses(ctx context.Context, userID string, startDate string, endDate string, limit int) ([]models.ExpensesReturn, error) {
	expenses, err := s.repo.GetExpensesByDateRange(ctx, userID, repositories.ExpenseFilters{
		StartDate: &startDate,
		EndDate:   &endDate,
	})
	if err != nil {
		return nil, err
	}

	// Convert to ExpensesReturn and sort by amount
	var expensesReturn []models.ExpensesReturn
	for _, expense := range expenses {
		expensesReturn = append(expensesReturn, models.ExpensesReturn{
			ExpenseId: expense.ExpenseId,
			Date:      expense.Date.Format("2006-01-02"),
			Name:      expense.Name,
			Amount:    expense.Amount,
			UserId:    expense.UserId,
			Status:    expense.Status,
		})
	}

	// Sort by amount (descending)
	sort.Slice(expensesReturn, func(i, j int) bool {
		return expensesReturn[i].Amount > expensesReturn[j].Amount
	})

	// Return top N expenses
	if len(expensesReturn) > limit {
		return expensesReturn[:limit], nil
	}
	return expensesReturn, nil
}