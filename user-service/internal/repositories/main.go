package repositories

import (
	"context"
	"errors"
	"fmt"
	"log"

	"user-service/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

var ErrUserNotFound = errors.New("ERROR : user not found")
var ErrUnespectedError = errors.New("unespected error")

// Define DynamoDB repository struct

type DynamoDBUsersRepository struct {
	client     *dynamodb.Client
	tableUsers string
}

func NewDynamoDBUsersRepository(client *dynamodb.Client, tableUsers string) *DynamoDBUsersRepository {
	return &DynamoDBUsersRepository{
		client:     client,
		tableUsers: tableUsers,
	}
}

// function to get user info by ID

func (c *DynamoDBUsersRepository) GetUserById(id string) (models.User, error) {
	//getUser(userId string, table string, c *dynamodb.Client) User {
	// call Getitem func
	output := models.User{}
	response, err := c.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(c.tableUsers),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberN{Value: id},
		},
	})
	if err != nil {
		log.Printf("unable to get item, %v", err)
		return output, err
	}
	if len(response.Item) == 0 {
		log.Printf("user not found")
		return models.User{}, ErrUserNotFound
	}
	// unmarshal item to models.user struct
	err2 := attributevalue.UnmarshalMap(response.Item, &output)
	if err2 != nil {
		log.Print("failed to unmarshal Items, %w", err2)
		return output, ErrUnespectedError
	}
	return output, nil
}

// function to get user by email
func (c *DynamoDBUsersRepository) GetUserByEmail(email string) (models.User, error) {
	output := models.User{}

	// create keycondition dynamodb expression for the query
	keyCondition := expression.Key("email").Equal(expression.Value(email))

	// create expression builder for the keyCondition
	builder := expression.NewBuilder().WithKeyCondition(keyCondition)
	expr, err := builder.Build()
	// Create the input for the dynamodb query
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(c.tableUsers),
		IndexName:                 aws.String("by_email"),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}
	// Execute the query

	response, err := c.client.Query(context.TODO(), queryInput)
	if err != nil {
		log.Printf("unable to query, %v", err)
		return models.User{}, err
	}

	if len(response.Items) == 0 {
		log.Print("user not found")
		return models.User{}, ErrUserNotFound
	}
	//unmarshall dynamodb output
	err2 := attributevalue.UnmarshalMap(response.Items[0], &output)
	if err2 != nil {
		log.Print("failed to unmarshal Items, %w", err2)
		return output, ErrUnespectedError
	}
	return output, nil

}

func (c *DynamoDBUsersRepository) CreateUser(u models.User) error {
	fmt.Println("repository createUser")
	item, err := attributevalue.MarshalMap(u)
	if err != nil {
		log.Printf("unable to marshal user, %v", err)
		return err
	}
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(c.tableUsers),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(email)"),
	}
	// call dynamodb put item
	_, err = c.client.PutItem(context.TODO(), input)

	if err != nil {
		log.Printf("error in DynamoDB put item, %v", err)
		return err
	}
	return nil
}
