package repositories

import (
	"context"
	"expenses-service/internal/common"
	"expenses-service/internal/models"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Define DynamoDB repository struct

type DynamoDBExpensesRepository struct {
	client        *dynamodb.Client
	expensesTable string
}

func NewDynamoDBExpensesRepository(client *dynamodb.Client, expensesTable string) *DynamoDBExpensesRepository {
	return &DynamoDBExpensesRepository{
		client:        client,
		expensesTable: expensesTable,
	}
}

// Function to Create a new expense

func (r *DynamoDBExpensesRepository) CreateExpense(expense models.Expense) (models.Expense, error) {
	// Create a new expense item
	item, err := attributevalue.MarshalMap(expense)
	if err != nil {
		log.Println("unable to marshal expense item", err)
		return models.Expense{}, err
	}

	// Create a new expense item
	_, err = r.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(r.expensesTable),
		Item:      item,
	})
	if err != nil {
		log.Println(common.ErrExpenseNoCreated, err)
		return models.Expense{}, common.ErrExpenseNoCreated
	}

	return models.Expense{}, nil
}

// Function to get a expense by id

func (r *DynamoDBExpensesRepository) GetExpenseById(id string) (models.Expense, error) {
	// Get expense item by id
	item, err := r.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(r.expensesTable),
		Key: map[string]types.AttributeValue{
			"id": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		log.Println(err)
		return models.Expense{}, err
	}
	if len(item.Item) == 0 {
		log.Println(common.ErrExpenseNotFound, err)
		return models.Expense{}, common.ErrExpenseNotFound
	}

	// Unmarshal the expense item
	var expense models.Expense
	err = attributevalue.UnmarshalMap(item.Item, &expense)
	if err != nil {
		log.Println(err)
		return models.Expense{}, err
	}

	return expense, nil
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
		log.Println(err)
		return err
	}

	return nil
}

// Function get expense by userID
func (r *DynamoDBExpensesRepository) GetExpensesByUserId(userId string) ([]models.Expense, error) {
	// create keycondition for userId
	keyCondition := expression.Key("userId").Equal(expression.Value(userId))

	// create expression for userId
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		log.Println("unable to build expression", err)
		return nil, err
	}
	// Get expenses by userID
	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.expensesTable),
		IndexName:                 aws.String("by_userid_and_category"),
		ExpressionAttributeNames:  expr.Names(),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeValues: expr.Values(),
	}

	items, err := r.client.Query(context.TODO(), input)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	// Unmarshal the expense items
	var expenses []models.Expense
	err = attributevalue.UnmarshalListOfMaps(items.Items, &expenses)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return expenses, nil
}

// funtion to get expenses filtered by category and userID as a global secondary indexes
func (r *DynamoDBExpensesRepository) GetExpensesByCategory(tag string, userId string) ([]models.Expense, error) {
	// create keycondition for tag and userId
	keyCondition := expression.Key("category").Equal(expression.Value(tag)).And(expression.Key("userId").Equal(expression.Value(userId)))

	// create expression for tag and userId
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		log.Println("unable to build expression", err)
		return nil, err
	}

	input := &dynamodb.QueryInput{
		TableName:                 aws.String(r.expensesTable),
		IndexName:                 aws.String("by_userid_and_category"),
		ExpressionAttributeNames:  expr.Names(),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeValues: expr.Values(),
	}

	response, err := r.client.Query(context.TODO(), input)

	if err != nil {
		return nil, err
	}

	// Unmarshal the expense response
	var expenses []models.Expense
	err = attributevalue.UnmarshalListOfMaps(response.Items, &expenses)
	if err != nil {
		return nil, err
	}
	return expenses, nil
}
