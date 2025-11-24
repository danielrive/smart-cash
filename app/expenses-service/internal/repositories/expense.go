package repositories

import (
	"context"
	"log/slog"
	"smart-cash/expenses-service/internal/common"
	"smart-cash/expenses-service/models"
	"smart-cash/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type DynamoDBExpensesRepository struct {
	client        *dynamodb.Client
	expensesTable string
	logger        *slog.Logger
}

func NewDynamoDBExpensesRepository(client *dynamodb.Client, expensesTable string, logger *slog.Logger) *DynamoDBExpensesRepository {
	return &DynamoDBExpensesRepository{
		client:        client,
		expensesTable: expensesTable,
		logger:        logger,
	}
}

// Function to Create a new expense

func (r *DynamoDBExpensesRepository) CreateExpense(ctx context.Context, expense models.Expense) (models.ExpensesReturn, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryCreateExpense", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("PutItem"),
		semconv.DBNameKey.String(r.expensesTable),
		attribute.String("db.table", r.expensesTable),
		attribute.String("expense.id", expense.ExpenseId),
		attribute.String("user.id", expense.UserId),
	)
	defer endSpan()

	output := models.ExpensesReturn{}

	r.logger.Debug("creating expense in database",
		slog.String("expense_id", expense.ExpenseId),
		slog.String("user_id", expense.UserId),
		slog.String("component", "repository"),
	)

	item, err := attributevalue.MarshalMap(expense)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("error marshaling expense item",
			slog.String("error", err.Error()),
			slog.String("expense_id", expense.ExpenseId),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}

	// Create a new expense item
	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.expensesTable),
		Item:      item,
	})
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("dynamodb error putting expense item",
			slog.String("error", err.Error()),
			slog.String("expense_id", expense.ExpenseId),
			slog.String("component", "repository"),
		)
		return output, common.ErrExpenseNotCreated
	}

	utils.AddSpanEvent(ctx, "expense created successfully",
		attribute.String("expense.id", expense.ExpenseId),
		attribute.Bool("db.item_created", true),
	)
	utils.SetSpanStatus(ctx, codes.Ok, "expense created successfully")

	r.logger.Info("expense created in database",
		slog.String("expense_id", expense.ExpenseId),
		slog.String("component", "repository"),
	)

	return createExpenserReturn(expense), nil
}

// Function to update expense
func (r *DynamoDBExpensesRepository) UpdateExpenseStatus(ctx context.Context, expense models.Expense) (models.ExpensesReturn, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryUpdateExpenseStatus", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("UpdateItem"),
		semconv.DBNameKey.String(r.expensesTable),
		attribute.String("db.table", r.expensesTable),
		attribute.String("expense.id", expense.ExpenseId),
	)
	defer endSpan()

	update := expression.Set(expression.Name("status"), expression.Value(expense.Status))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("dynamodb update expression couldn't be created",
			"error", err.Error(),
			"expenseId", expense.ExpenseId,
		)
		return models.ExpensesReturn{}, common.ErrInternalError
	}
	// Define the key of the item to update
	expId, err := attributevalue.Marshal(expense.ExpenseId)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("dynamodb udpate key couldn't be created",
			"error", err.Error(),
			"expenseId", expense.ExpenseId,
		)
		return models.ExpensesReturn{}, common.ErrInternalError
	}

	inputUpdate := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.expensesTable),
		Key:                       map[string]types.AttributeValue{"expenseId": expId},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	}
	response, err := r.client.UpdateItem(ctx, inputUpdate)

	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("status couldn't be updated",
			"error", err.Error(),
			"expenseId", expense.ExpenseId,
		)
		return models.ExpensesReturn{}, common.ErrInternalError
	}
	// Unmarshal the response
	if response.Attributes != nil {
		for _, value := range response.Attributes {
			// Declare a variable to hold the unmarshaled value
			var readableValue interface{}
			// Unmarshal the AttributeValue into a generic interface{}
			err := attributevalue.Unmarshal(value, &readableValue)
			if err != nil {
				r.logger.Error("failed to unmarshal attribute value",
					"error", err.Error(),
					"expenseId", expense.ExpenseId,
				)
			}
			expense.Status = readableValue.(string)

		}
	} else {
		r.logger.Info("no attributes returned")
	}

	utils.AddSpanEvent(ctx, "expense status updated successfully",
		attribute.String("expense.id", expense.ExpenseId),
		attribute.String("expense.status", expense.Status),
		attribute.Bool("db.item_updated", true),
	)
	utils.SetSpanStatus(ctx, codes.Ok, "expense status updated successfully")

	return createExpenserReturn(expense), nil

}

// Function to get a expense by id

func (r *DynamoDBExpensesRepository) GetExpenseById(ctx context.Context, id string) (models.Expense, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryGetExpenseById", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("GetItem"),
		semconv.DBNameKey.String(r.expensesTable),
		attribute.String("db.table", r.expensesTable),
		attribute.String("expense.id", id),
	)
	defer endSpan()

	output := models.Expense{}
	// Get expense item by id
	item, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.expensesTable),
		Key: map[string]types.AttributeValue{
			"expenseId": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("dynamodb couldn't get the item",
			"error", err.Error(),
			"expenseId", id,
		)
		return output, common.ErrInternalError
	}
	if len(item.Item) == 0 {
		utils.AddSpanEvent(ctx, "expense not found", attribute.Bool("db.item_found", false))
		utils.SetSpanStatus(ctx, codes.Ok, "expense not found")
		r.logger.Debug("expense not found in database",
			slog.String("expense_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrExpenseNotFound
	}

	// Unmarshal the expense item
	err = attributevalue.UnmarshalMap(item.Item, &output)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("error unmarshaling expense item",
			slog.String("error", err.Error()),
			slog.String("expense_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}

	utils.AddSpanEvent(ctx, "expense retrieved successfully",
		attribute.Bool("db.item_found", true),
		attribute.String("expense.id", output.ExpenseId),
	)
	utils.SetSpanStatus(ctx, codes.Ok, "expense retrieved successfully")

	r.logger.Debug("expense retrieved from database",
		slog.String("expense_id", id),
		slog.String("component", "repository"),
	)

	return output, nil
}

// Function get expense by userID
func (r *DynamoDBExpensesRepository) GetExpByUserIdorCat(ctx context.Context, k string, v string) ([]models.Expense, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryGetExpByUserIdorCat", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("Query"),
		semconv.DBNameKey.String(r.expensesTable),
		attribute.String("db.table", r.expensesTable),
		attribute.String("db.index", "by_"+k),
		attribute.String("query.key", k),
		attribute.String("query.value", v),
	)
	defer endSpan()

	// create keycondition for userId
	output := []models.Expense{}

	keyCondition := expression.Key(k).Equal(expression.Value(v))

	// create expression
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("dynamodb expression couldn't be created",
			"error", err.Error(),
		)
		return output, common.ErrInternalError
	}
	// Get expenses by userID
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(r.expensesTable),
		IndexName:                 aws.String("by_" + k),
		ExpressionAttributeNames:  expr.Names(),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeValues: expr.Values(),
	}

	response, err := r.client.Query(ctx, queryInput)

	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("dynamodb query failed",
			"error", err.Error(),
		)
		return output, common.ErrInternalError
	}

	if len(response.Items) == 0 {
		utils.AddSpanEvent(ctx, "expenses not found", attribute.Bool("db.items_found", false))
		utils.SetSpanStatus(ctx, codes.Ok, "expenses not found")
		r.logger.Info("expense not found",
			"tag", k,
			"value", v,
		)
		return output, common.ErrExpenseNotFound
	}

	// Unmarshal the expense items
	err = attributevalue.UnmarshalListOfMaps(response.Items, &output)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("failed to unmarshal attribute value",
			"error", err.Error(),
		)
		return []models.Expense{}, common.ErrInternalError
	}

	utils.AddSpanEvent(ctx, "expenses retrieved successfully",
		attribute.Bool("db.items_found", true),
		attribute.Int("db.items_count", len(output)),
	)
	utils.SetSpanStatus(ctx, codes.Ok, "expenses retrieved successfully")

	return output, nil
}

// Function to delete a expense by id

func (r *DynamoDBExpensesRepository) DeleteExpenseById(ctx context.Context, id string) error {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryDeleteExpenseById", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("DeleteItem"),
		semconv.DBNameKey.String(r.expensesTable),
		attribute.String("db.table", r.expensesTable),
		attribute.String("expense.id", id),
	)
	defer endSpan()

	// Delete expense item by id
	_, err := r.client.DeleteItem(ctx, &dynamodb.DeleteItemInput{
		TableName: aws.String(r.expensesTable),
		Key: map[string]types.AttributeValue{
			"expenseId": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("dynamodb delete operation failed",
			"error", err.Error(),
		)
		return common.ErrInternalError
	}

	utils.AddSpanEvent(ctx, "expense deleted successfully", attribute.Bool("db.item_deleted", true))
	utils.SetSpanStatus(ctx, codes.Ok, "expense deleted successfully")

	return nil
}

func createExpenserReturn(expense models.Expense) models.ExpensesReturn {
	return models.ExpensesReturn{
		Date:        expense.Date,
		ExpenseId:   expense.ExpenseId,
		Name:        expense.Name,
		Description: expense.Description,
		Status:      expense.Status,
		Amount:      expense.Amount,
		Category:    expense.Category,
		Tags:        expense.Tags,
	}
}
