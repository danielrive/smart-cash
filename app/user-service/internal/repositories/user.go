package repositories

import (
	"context"
	"log"
	"smart-cash/user-service/internal/common"
	"smart-cash/user-service/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// define UUID interface
type UUIDHelper interface {
	New() string
}

// Define DynamoDB repository struct
type DynamoDBUsersRepository struct {
	client     *dynamodb.Client
	tableUsers string
	uuid       UUIDHelper
}

func NewDynamoDBUsersRepository(client *dynamodb.Client, tableUsers string, uuid UUIDHelper) *DynamoDBUsersRepository {
	return &DynamoDBUsersRepository{
		client:     client,
		tableUsers: tableUsers,
		uuid:       uuid,
	}
}

// Function to Get user by ID
func (c *DynamoDBUsersRepository) GetUserById(id string) (models.UserResponse, error) {
	output := models.UserResponse{}

	// create input for get item
	input := &dynamodb.GetItemInput{
		TableName: aws.String(c.tableUsers),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{
				Value: id,
			},
		},
	}
	// call dynamoDB GetItem
	response, err := c.client.GetItem(context.TODO(), input)

	if err != nil {
		log.Println(common.ErrInternalError, err)
		return output, err
	}
	if len(response.Item) == 0 {
		log.Println(common.ErrUserNotFound, err)
		return output, common.ErrUserNotFound
	}
	// unmarshal item to models.user struct
	err2 := attributevalue.UnmarshalMap(response.Item, &output)
	if err2 != nil {
		log.Println(common.ErrInternalError, err2)
		return output, common.ErrInternalError
	}
	return output, nil
}

// Function to Create user

func (c *DynamoDBUsersRepository) CreateUser(u models.User) (models.UserResponse, error) {

	output := models.UserResponse{}
	u.UserId = c.uuid.New()

	item, err := attributevalue.MarshalMap(u)
	if err != nil {
		log.Println(common.ErrInternalError, err)
		return output, err
	}
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(c.tableUsers),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(userId)"),
	}
	// call dynamodb put item
	_, err = c.client.PutItem(context.TODO(), input)

	if err != nil {
		log.Println(common.ErrUserNoCreated, err)
		return output, common.ErrUserNoCreated
	}
	// create output response
	output.UserId = u.UserId
	output.Username = u.Username
	output.Email = u.Email
	output.Active = u.Active
	return output, nil
}

// Function to Update User

func (c *DynamoDBUsersRepository) UpdateUser(u models.User) (models.UserResponse, error) {
	output := models.UserResponse{}

	item, err := attributevalue.MarshalMap(u)
	if err != nil {
		log.Println(err)
		return output, err
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(c.tableUsers),
		Item:      item,
	}
	// call dynamodb put item
	_, err = c.client.PutItem(context.TODO(), input)

	if err != nil {
		log.Println(common.ErrUserNoCreated, err)
		return output, common.ErrUserNoCreated
	}
	// create output response
	output.UserId = u.UserId
	output.Username = u.Username
	output.Active = u.Active

	return output, nil
}

// Function to Get user by email
func (c *DynamoDBUsersRepository) GetUserByEmailorUsername(k string, v string) (models.UserResponse, error) {
	output := models.UserResponse{}

	// create keycondition dynamodb expression for the query
	keyCondition := expression.Key(k).Equal(expression.Value(v))

	// create expression builder for the keyCondition
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()

	if err != nil {
		log.Println(common.ErrInternalError, err)
		return output, common.ErrInternalError
	}

	// Create the input for the dynamodb query
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(c.tableUsers),
		IndexName:                 aws.String("by_" + k),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}
	// Execute the query

	response, err := c.client.Query(context.TODO(), queryInput)

	if err != nil {
		log.Println(common.ErrInternalError, err)
		return output, common.ErrInternalError
	}

	if len(response.Items) == 0 {
		log.Println(common.ErrUserNotFound, err)
		return output, common.ErrUserNotFound
	}
	//unmarshall dynamodb output
	err2 := attributevalue.UnmarshalMap(response.Items[0], &output)
	if err2 != nil {
		log.Println(common.ErrInternalError, err2)
		return models.UserResponse{}, common.ErrInternalError
	}
	return output, nil

}
