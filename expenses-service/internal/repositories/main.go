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
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Define error types
var ErrExpenseNotFound = errors.New("ERROR : expenses not found")
var ErrExpenseNotCreated = errors.New("ERROR : expense not created")
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
		return models.Expense{}, ErrUnespectedError
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

<<<<<<< HEAD
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

	// Execute the query

	response, err := r.client.Query(context.TODO(), queryInput)

	if err != nil {
		log.Printf("unable to query, %v", err)
		return []models.Expense{}, ErrUnespectedError
	}

	if len(response.Items) == 0 {
		log.Print("expenses not found")
		return []models.Expense{}, ErrExpenseNotFound
	}

	//unmarshall dynamodb output
	err2 := attributevalue.UnmarshalListOfMaps(response.Items, &output)
	if err2 != nil {
		log.Print("failed to unmarshal Items, %w", err2)
		return output, ErrUnespectedError
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
// 		return output, ErrUnespectedError
// 	}
// 	return output, nil
// }
=======
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
		log.Println("unable to get expense", err)
		return models.Expense{}, err
	}
	if len(item.Item) == 0 {
		log.Printf("user not found")
		return models.Expense{}, ErrExpenseNotFound
	}

	// Unmarshal the expense item
	var expense models.Expense
	err = attributevalue.UnmarshalMap(item.Item, &expense)
	if err != nil {
		log.Println("unable to unmarshal expense item", err)
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
		log.Println("unable to delete expense", err)
		return err
	}

	return nil
}

// Function get expense by userID
func (r *DynamoDBExpensesRepository) GetExpenseByUserId(userId string) ([]models.Expense, error) {
	// create keycondition for userId
	keyCondition := expression.Key("userId").Equal(expression.Value(userId))

	// create expression for userId
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		log.Println("unable to build expression", err)
		return nil, err
	}

	// Get expense items by userID
	items, err := r.client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 aws.String(r.expensesTable),
		IndexName:                 aws.String("by_userID"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeValues: expr.Values(),
	})

	if err != nil {
		log.Println("unable to get expenses", err)
		return nil, err
	}

	// Unmarshal the expense items
	var expenses []models.Expense
	err = attributevalue.UnmarshalListOfMaps(items.Items, &expenses)
	if err != nil {
		log.Println("unable to unmarshal expense items", err)
		return nil, err
	}

	return expenses, nil
}

// funtion to get expenses filtered by tag and userID as a global secondary indexes
func (r *DynamoDBExpensesRepository) GetExpensesByTag(tag string, userId string) ([]models.Expense, error) {
	// create keycondition for tag and userId
	keyCondition := expression.Key("tag").Equal(expression.Value(tag)).And(expression.Key("userId").Equal(expression.Value(userId)))

	// create expression for tag and userId
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()
	if err != nil {
		log.Println("unable to build expression", err)
		return nil, err
	}

	// Get expense items by tag and userId
	items, err := r.client.Query(context.TODO(), &dynamodb.QueryInput{
		TableName:                 aws.String(r.expensesTable),
		IndexName:                 aws.String("by_tag"),
		ExpressionAttributeNames:  expr.Names(),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeValues: expr.Values(),
	})

	if err != nil {
		log.Println("unable to get expenses", err)
		return nil, err
	}

	// Unmarshal the expense items
	var expenses []models.Expense
	err = attributevalue.UnmarshalListOfMaps(items.Items, &expenses)
	if err != nil {
		log.Println("unable to unmarshal expense items", err)
		return nil, err
	}
	return expenses, nil
}
>>>>>>> 2826218 (update k8 version to 1.29)
