package repositories

import (
	"context"
	"log/slog"
	"smart-cash/expenses-service/internal/common"

	"smart-cash/expenses-service/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.opentelemetry.io/otel"
)

// Define DynamoDB repository struct
type UUIDHelper interface {
	New() string
}

type DynamoDBExpensesRepository struct {
	client        *dynamodb.Client
	expensesTable string
	uuid          UUIDHelper
	logger        *slog.Logger
}

func NewDynamoDBExpensesRepository(client *dynamodb.Client, expensesTable string, uuid UUIDHelper, logger *slog.Logger) *DynamoDBExpensesRepository {
	return &DynamoDBExpensesRepository{
		client:        client,
		expensesTable: expensesTable,
		uuid:          uuid,
		logger:        logger,
	}
}

// Function to Create a new expense

func (r *DynamoDBExpensesRepository) CreateExpense(ctx context.Context, expense models.Expense) (models.ExpensesReturn, error) {
	tr := otel.Tracer("expenses-service")
	_, childSpan := tr.Start(ctx, "RepositoryCreateExpense")
	defer childSpan.End()
	// Create a new expense item
	output := models.ExpensesReturn{}
	expense.ExpenseId = r.uuid.New()
	item, err := attributevalue.MarshalMap(expense)
	if err != nil {
		r.logger.Error("error while unmarshaling DynamoDB item",
			"error", err.Error(),
			"expenseId", expense.ExpenseId,
		)
		return output, common.ErrInternalError
	}

	// Create a new expense item
	_, err = r.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(r.expensesTable),
		Item:      item,
	})
	if err != nil {
		r.logger.Error("dynamodb error while putting item",
			"error", err.Error(),
			"expenseId", expense.ExpenseId,
		)
		return output, common.ErrExpenseNoCreated
	}
	return createExpenserReturn(expense), nil
}

// Function to update expense
func (r *DynamoDBExpensesRepository) UpdateExpenseStatus(ctx context.Context, expense models.Expense) (models.ExpensesReturn, error) {
	tr := otel.Tracer("expenses-service")
	_, childSpan := tr.Start(ctx, "RepositoryUpdateExpenseStatus")
	defer childSpan.End()

	update := expression.Set(expression.Name("status"), expression.Value(expense.Status))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		r.logger.Error("dynamodb update expression couldn't be created",
			"error", err.Error(),
			"expenseId", expense.ExpenseId,
		)
		return models.ExpensesReturn{}, common.ErrInternalError
	}
	// Define the key of the item to update
	expId, err := attributevalue.Marshal(expense.ExpenseId)
	if err != nil {
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
	response, err := r.client.UpdateItem(context.TODO(), inputUpdate)

	if err != nil {
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

	return createExpenserReturn(expense), nil

}

// Function to get a expense by id

func (r *DynamoDBExpensesRepository) GetExpenseById(ctx context.Context, id string) (models.Expense, error) {
	tr := otel.Tracer("expenses-service")
	_, childSpan := tr.Start(ctx, "RepositoryGetExpenseById")
	defer childSpan.End()

	output := models.Expense{}
	// Get expense item by id
	item, err := r.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(r.expensesTable),
		Key: map[string]types.AttributeValue{
			"expenseId": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		r.logger.Error("dynamodb couldn't get the item",
			"error", err.Error(),
			"expenseId", id,
		)
		return output, common.ErrInternalError
	}
	if len(item.Item) == 0 {
		r.logger.Info("expense not found",
			"expenseId", id,
		)
		return output, common.ErrExpenseNotFound
	}

	// Unmarshal the expense item
	err = attributevalue.UnmarshalMap(item.Item, &output)
	if err != nil {
		r.logger.Error("failed to unmarshal attribute value",
			"error", err.Error(),
			"expenseId", id,
		)
		return output, common.ErrInternalError
	}

	return output, nil
}

// Function get expense by userID
func (r *DynamoDBExpensesRepository) GetExpByUserIdorCat(ctx context.Context, k string, v string) ([]models.Expense, error) {
	tr := otel.Tracer("expenses-service")
	_, childSpan := tr.Start(ctx, "RepositoryGetExpByUserIdorCat")
	defer childSpan.End()

	// create keycondition for userId
	output := []models.Expense{}

	keyCondition := expression.Key(k).Equal(expression.Value(v))

	// create expression
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
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

	response, err := r.client.Query(context.TODO(), queryInput)

	if err != nil {
		r.logger.Error("dynamodb query failed",
			"error", err.Error(),
		)
		return output, common.ErrInternalError
	}

	if len(response.Items) == 0 {
		r.logger.Info("expense not found",
			"tag", k,
			"value", v,
		)
		return output, common.ErrExpenseNotFound
	}

	// Unmarshal the expense items
	err = attributevalue.UnmarshalListOfMaps(response.Items, &output)
	if err != nil {
		r.logger.Error("failed to unmarshal attribute value",
			"error", err.Error(),
		)
		return []models.Expense{}, common.ErrInternalError
	}

	return output, nil
}

// Function to delete a expense by id

func (r *DynamoDBExpensesRepository) DeleteExpenseById(ctx context.Context, id string) error {
	tr := otel.Tracer("expenses-service")
	_, childSpan := tr.Start(ctx, "RepositoryDeleteExpenseById")
	defer childSpan.End()

	// Delete expense item by id
	_, err := r.client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(r.expensesTable),
		Key: map[string]types.AttributeValue{
			"expenseId": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		r.logger.Error("dynamodb delete operation failed",
			"error", err.Error(),
		)
		return common.ErrInternalError
	}

	return nil
}

func createExpenserReturn(expense models.Expense) models.ExpensesReturn {
	return models.ExpensesReturn{
		Date:      expense.Date,
		ExpenseId: expense.ExpenseId,
		Name:      expense.Name,
		Status:    expense.Status,
		Amount:    expense.Amount,
	}
}
