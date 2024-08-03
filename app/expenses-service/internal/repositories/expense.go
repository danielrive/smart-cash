package repositories

import (
	"context"
	"fmt"
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
		log.Printf("internal error while unmarshaling DynamoDB item %v:", err)
		return output, common.ErrInternalError
	}

	// Create a new expense item
	_, err = r.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(r.expensesTable),
		Item:      item,
	})
	if err != nil {
		log.Printf("dynamodb error while putting item %v:", err)
		return output, common.ErrExpenseNoCreated
	}
	return createExpenserReturn(expense), nil
}

// Function to update expense
func (r *DynamoDBExpensesRepository) UpdateExpenseStatus(expense models.Expense) (models.ExpensesReturn, error) {

	update := expression.Set(expression.Name("status"), expression.Value(expense.Status))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		log.Printf("dynamodb udpate expression couldn't be created: %v", err)
		return models.ExpensesReturn{}, common.ErrInternalError
	}
	// Define the key of the item to update
	expId, err := attributevalue.Marshal(expense.ExpenseId)
	if err != nil {
		log.Printf("dynamodb udpate key couldn't be created: %v", err)
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
		log.Printf("status for expense %v couldn't be updated %v:", expense.ExpenseId, err)
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
				log.Printf("failed to unmarshal attribute value: %s", err)
			}
			expense.Status = readableValue.(string)
			// Print the attribute name and its readable value
		}
	} else {
		fmt.Println("No attributes returned.")
	}

	return createExpenserReturn(expense), nil

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

func createExpenserReturn(expense models.Expense) models.ExpensesReturn {
	return models.ExpensesReturn{
		Date:      expense.Date,
		ExpenseId: expense.ExpenseId,
		Name:      expense.Name,
		Status:    expense.Status,
		Amount:    expense.Amount,
	}
}
