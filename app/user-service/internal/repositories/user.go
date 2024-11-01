package repositories

import (
	"context"
	"log/slog"
	"smart-cash/user-service/internal/common"
	"smart-cash/user-service/models"

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
	logger     *slog.Logger
}

func NewDynamoDBUsersRepository(client *dynamodb.Client, tableUsers string, uuid UUIDHelper, logger *slog.Logger) *DynamoDBUsersRepository {
	return &DynamoDBUsersRepository{
		client:     client,
		tableUsers: tableUsers,
		uuid:       uuid,
		logger:     logger,
	}
}

// Function to Get user by ID
func (r *DynamoDBUsersRepository) GetUserById(id string) (models.UserResponse, error) {
	output := models.UserResponse{}

	// create input for get item
	input := &dynamodb.GetItemInput{
		TableName: aws.String(r.tableUsers),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{
				Value: id,
			},
		},
	}
	// call dynamoDB GetItem
	response, err := r.client.GetItem(context.TODO(), input)

	if err != nil {
		r.logger.Error("dynamodb get item failed",
			"error", err.Error(),
			"userId", id,
		)
		return output, err
	}
	if len(response.Item) == 0 {
		r.logger.Error("user not found",
			"userId", id,
		)
		return output, common.ErrUserNotFound
	}
	// unmarshal item to models.user struct
	err = attributevalue.UnmarshalMap(response.Item, &output)
	if err != nil {
		r.logger.Error("error unmashaling map",
			"error", err.Error(),
			"userId", id,
		)
		return output, common.ErrInternalError
	}
	return output, nil
}

// Function to Create user

func (r *DynamoDBUsersRepository) CreateUser(u models.User) (models.UserResponse, error) {

	output := models.UserResponse{}
	u.UserId = r.uuid.New()

	item, err := attributevalue.MarshalMap(u)
	if err != nil {
		r.logger.Error("error marshaling map",
			"error", err.Error(),
			"userId", u.UserId,
		)
		return output, err
	}
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableUsers),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(userId)"),
	}
	// call dynamodb put item
	_, err = r.client.PutItem(context.TODO(), input)

	if err != nil {
		r.logger.Error("dynamodb error put item",
			"error", err.Error(),
			"userId", u.UserId,
		)
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

func (r *DynamoDBUsersRepository) UpdateUser(u models.User) (models.UserResponse, error) {
	output := models.UserResponse{}

	item, err := attributevalue.MarshalMap(u)
	if err != nil {
		r.logger.Error("error marshaling map",
			"error", err.Error(),
			"userId", u.UserId,
		)
		return output, err
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableUsers),
		Item:      item,
	}
	// call dynamodb put item
	_, err = r.client.PutItem(context.TODO(), input)

	if err != nil {
		r.logger.Error("dynamodb error put item",
			"error", err.Error(),
			"userId", u.UserId,
		)
		return output, common.ErrUserNoCreated
	}
	// create output response
	output.UserId = u.UserId
	output.Username = u.Username
	output.Active = u.Active

	return output, nil
}

// Function to Get user by email
func (r *DynamoDBUsersRepository) GetUserByEmailorUsername(k string, v string) (models.User, error) {
	output := models.User{}

	// create keycondition dynamodb expression for the query
	keyCondition := expression.Key(k).Equal(expression.Value(v))

	// create expression builder for the keyCondition
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()

	if err != nil {
		r.logger.Error("dynamodb error building expression",
			"error", err.Error(),
			k, v,
		)
		return output, common.ErrInternalError
	}

	// Create the input for the dynamodb query
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(r.tableUsers),
		IndexName:                 aws.String("by_" + k),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}
	// Execute the query

	response, err := r.client.Query(context.TODO(), queryInput)

	if err != nil {
		r.logger.Error("dynamodb error query item",
			"error", err.Error(),
			k, v,
		)
		return output, common.ErrInternalError
	}

	if len(response.Items) == 0 {
		r.logger.Error("user not found",
			k, v,
		)
		return output, common.ErrUserNotFound
	}
	//unmarshall dynamodb output
	err = attributevalue.UnmarshalMap(response.Items[0], &output)
	if err != nil {
		r.logger.Error("error unmashaling map",
			"error", err.Error(),
			k, v,
		)
		return output, common.ErrInternalError
	}
	return output, nil

}
