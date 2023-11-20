package repositories

import (
	"context"
	"errors"
	"expenses-service/internal/models"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Define error types
var ErrExpenseNotFound = errors.New("ERROR : expenses not found")
var ErrExpenseNotCreated = errors.New("ERROR : expense not created")
var ErrInternalError = errors.New("ERROR: internal error")

// Define DynamoDB repository struct

type DynamoDBExpensesRepository struct {
	client        *dynamodb.Client
	expensesTable string
}

// Define DynamoDB repository methods
// func to create a new expense

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
		return models.Expense{}, ErrInternalError
	}

	// Create a new expense item
	_, err = r.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(r.expensesTable),
		Item:      item,
	})
	if err != nil {
		log.Println("unable to put expense item", err)
		return models.Expense{}, ErrExpenseNotCreated
	}

	return models.Expense{}, nil
}

// Function to Get expenses by userId and category

func (r *DynamoDBExpensesRepository) GetExpensesByUserIdAndCategory(userId string, category string) ([]models.Expense, error) {
	output := []models.Expense{}

	// create expression to filter by userId and category
	condition := expression.And(
		expression.Name("userId").Equal(expression.Value(userId)), expression.Name("category").Equal(expression.Value(category)),
	)

	builder := expression.NewBuilder().WithCondition(condition)

	expres, err := builder.Build()
	// Create a new expense item
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(r.expensesTable),
		KeyConditionExpression:    expres.Condition(),
		IndexName:                 aws.String("by_userid_and_category"),
		ExpressionAttributeNames:  expres.Names(),
		ExpressionAttributeValues: expres.Values(),
	}
	// validate if there is error in building the expression for dynamo
	if err != nil {
		log.Printf("unable to build dynamo expression, %v", err)
		return []models.Expense{}, ErrInternalError
	}

	// Execute the query

	response, err := r.client.Query(context.TODO(), queryInput)

	if err != nil {
		log.Printf("unable to query, %v", err)
		return []models.Expense{}, ErrInternalError
	}

	if len(response.Items) == 0 {
		log.Print("expenses not found")
		return []models.Expense{}, ErrExpenseNotFound
	}

	//unmarshall dynamodb output
	err2 := attributevalue.UnmarshalListOfMaps(response.Items, &output)
	if err2 != nil {
		log.Print("failed to unmarshal Items, %w", err2)
		return output, ErrInternalError
	}
	return output, nil
}

// Function to get expenses by userId

// func (r *DynamoDBExpensesRepository) GetExpensesByUserId(userId string) (models.Expense, error) {
// 	output := models.Expense{}

// 	// create expression to filter by userId and category
// 	keyCondition := expression.Key("userId").Equal(expression.Value(userId))

// 	builder := expression.NewBuilder().WithKeyCondition(keyCondition)

// 	expres, err := builder.Build()
// 	// Create a new expense item
// 	queryInput := &dynamodb.QueryInput{
// 		TableName:                 aws.String(r.expensesTable),
// 		KeyConditionExpression:    expres.KeyCondition(),
// 		IndexName:                 aws.String("by_userid_and_category"),
// 		ExpressionAttributeNames:  expres.Names(),
// 		ExpressionAttributeValues: expres.Values(),
// 	}

// 	// Execute the query

// 	response, err := r.client.Query(context.TODO(), queryInput)

// 	if err != nil {
// 		log.Printf("unable to query, %v", err)
// 		return models.Expense{}, err
// 	}

// 	if len(response.Items) == 0 {
// 		log.Print("expenses not found")
// 		return models.Expense{}, ErrExpenseNotFound
// 	}

// 	//unmarshall dynamodb output
// 	err2 := attributevalue.UnmarshalMap(response.Items[0], &output)
// 	if err2 != nil {
// 		log.Print("failed to unmarshal Items, %w", err2)
// 		return output, ErrInternalError
// 	}
// 	return output, nil
// }
