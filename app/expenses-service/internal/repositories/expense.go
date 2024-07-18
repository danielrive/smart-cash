package repositories

import (
	"context"
	"smart-cash/expenses-service/internal/common"

	"log"
	"smart-cash/expenses-service/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Define DynamoDB repository struct
type UUIDHelper interface {
	New() string
}

type DynamoDBExpensesRepository struct {
	client        *dynamodb.Client
	expensesTable string
	uuid          UUIDHelper
}

func NewDynamoDBExpensesRepository(client *dynamodb.Client, expensesTable string, uuid UUIDHelper) *DynamoDBExpensesRepository {
	return &DynamoDBExpensesRepository{
		client:        client,
		expensesTable: expensesTable,
		uuid:          uuid,
	}
}

// Function to Create a new expense

func (r *DynamoDBExpensesRepository) CreateExpense(expense models.Expense) (models.ExpensesReturn, error) {
	// Create a new expense item
	output := models.ExpensesReturn{}
	expense.ExpenseId = r.uuid.New()
	item, err := attributevalue.MarshalMap(expense)
	if err != nil {
		log.Println(common.ErrInternalError, err)
		return output, common.ErrInternalError
	}

	// Create a new expense item
	_, err = r.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(r.expensesTable),
		Item:      item,
	})
	if err != nil {
		log.Println(common.ErrExpenseNoCreated, err)
		return output, common.ErrExpenseNoCreated
	}
	output.Date = expense.Date
	output.ExpenseId = expense.ExpenseId
	output.Name = expense.Name

	return output, nil
}

// Function to get a expense by id

func (r *DynamoDBExpensesRepository) GetExpenseById(id string) (models.Expense, error) {
	output := models.Expense{}
	// Get expense item by id
	item, err := r.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(r.expensesTable),
		Key: map[string]types.AttributeValue{
			"expenseId": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		log.Println(common.ErrInternalError, err)
		return output, common.ErrInternalError
	}
	if len(item.Item) == 0 {
		log.Println(common.ErrExpenseNotFound, err)
		return output, common.ErrExpenseNotFound
	}

	// Unmarshal the expense item
	err = attributevalue.UnmarshalMap(item.Item, &output)
	if err != nil {
		log.Println(common.ErrInternalError, err)
		return output, common.ErrInternalError
	}

	return output, nil
}

// Function get expense by userID
func (r *DynamoDBExpensesRepository) GetExpByUserIdorCat(k string, v string) ([]models.Expense, error) {
	// create keycondition for userId
	output := []models.Expense{}

	keyCondition := expression.Key(k).Equal(expression.Value(v))

	// create expression
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		log.Println(common.ErrInternalError, err)
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
		log.Println(common.ErrInternalError, err)
		return output, common.ErrInternalError
	}

	if len(response.Items) == 0 {
		log.Println(common.ErrExpenseNotFound, err)
		return output, common.ErrExpenseNotFound
	}

	// Unmarshal the expense items
	err = attributevalue.UnmarshalListOfMaps(response.Items, &output)
	if err != nil {
		log.Println(common.ErrInternalError, err)
		return []models.Expense{}, common.ErrInternalError
	}

	return output, nil
}

// Function to delete a expense by id

func (r *DynamoDBExpensesRepository) DeleteExpenseById(id string) error {
	// Delete expense item by id
	_, err := r.client.DeleteItem(context.TODO(), &dynamodb.DeleteItemInput{
		TableName: aws.String(r.expensesTable),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		log.Println(common.ErrInternalError, err)
		return common.ErrInternalError
	}

	return nil
}
