package repositories

import (
	"context"
	"errors"
	"expenses-service/internal/models"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

// Define error types
var ErrExpenseNotFound = errors.New("ERROR : expense not found")
var ErrUnespectedError = errors.New("unespected error")

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
		return expense, err
	}

	// Create a new expense item
	_, err = r.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(r.expensesTable),
		Item:      item,
	})
	if err != nil {
		log.Println("unable to put expense item", err)
		return expense, err
	}

	return expense, nil
}
