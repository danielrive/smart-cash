package repositories

import (
	"context"
	"smart-cash/expenses-service/internal/common"
	"smart-cash/expenses-service/models"
	"smart-cash/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

// ExpenseFilters represents the filter options for querying expenses
type ExpenseFilters struct {
	StartDate *string // Format: YYYY-MM-DD
	EndDate   *string // Format: YYYY-MM-DD
	Category  *string
	Status    *string
	MinAmount *float64
	MaxAmount *float64
	Tags      []string
	Page      int
	PageSize  int
}

// GetExpensesByDateRange retrieves expenses within a date range
func (r *DynamoDBExpensesRepository) GetExpensesByDateRange(ctx context.Context, userID string, filters ExpenseFilters) ([]models.Expense, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryGetExpensesByDateRange", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("Query"),
		semconv.DBNameKey.String(r.expensesTable),
		attribute.String("db.table", r.expensesTable),
		attribute.String("user.id", userID),
	)
	defer endSpan()

	// Create the base key condition
	keyEx := expression.Key("userId").Equal(expression.Value(userID))

	// Add date range condition if provided
	var filterEx expression.ConditionBuilder
	var filterSet bool

	if filters.StartDate != nil && filters.EndDate != nil {
		filterEx = expression.Name("date").Between(
			expression.Value(*filters.StartDate),
			expression.Value(*filters.EndDate),
		)
		filterSet = true
	}

	// Add category filter if provided
	if filters.Category != nil {
		categoryFilter := expression.Name("category").Equal(expression.Value(*filters.Category))
		if filterSet {
			filterEx = filterEx.And(categoryFilter)
		} else {
			filterEx = categoryFilter
			filterSet = true
		}
	}

	// Add status filter if provided
	if filters.Status != nil {
		statusFilter := expression.Name("status").Equal(expression.Value(*filters.Status))
		if filterSet {
			filterEx = filterEx.And(statusFilter)
		} else {
			filterEx = statusFilter
			filterSet = true
		}
	}

	// Add amount range filters if provided
	if filters.MinAmount != nil {
		minAmountFilter := expression.Name("amount").GreaterThanEqual(expression.Value(*filters.MinAmount))
		if filterSet {
			filterEx = filterEx.And(minAmountFilter)
		} else {
			filterEx = minAmountFilter
			filterSet = true
		}
	}

	if filters.MaxAmount != nil {
		maxAmountFilter := expression.Name("amount").LessThanEqual(expression.Value(*filters.MaxAmount))
		if filterSet {
			filterEx = filterEx.And(maxAmountFilter)
		} else {
			filterEx = maxAmountFilter
			filterSet = true
		}
	}

	// Build the expression
	builder := expression.NewBuilder().WithKeyCondition(keyEx)
	if filterSet {
		builder = builder.WithFilter(filterEx)
	}

	expr, err := builder.Build()
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("error building expression for query",
			"error", err.Error(),
			"userId", userID,
		)
		return nil, err
	}

	// Prepare the query input
	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.expensesTable),
		KeyConditionExpression:    expr.KeyCondition(),
		FilterExpression:          expr.Filter(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}

	// Execute the query
	result, err := r.client.Query(ctx, input)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("error querying expenses",
			"error", err.Error(),
			"userId", userID,
		)
		return nil, err
	}

	// Unmarshal the results
	var expenses []models.Expense
	err = attributevalue.UnmarshalListOfMaps(result.Items, &expenses)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("error unmarshaling query results",
			"error", err.Error(),
			"userId", userID,
		)
		return nil, err
	}

	utils.AddSpanEvent(ctx, "expenses retrieved by date range",
		attribute.Int("db.items_count", len(expenses)),
		attribute.Bool("db.items_found", len(expenses) > 0),
	)

	return expenses, nil
}
