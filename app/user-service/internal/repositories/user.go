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
	"go.opentelemetry.io/otel"
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
func (r *DynamoDBUsersRepository) GetUserById(ctx context.Context, id string) (models.UserResponse, error) {
	tr := otel.Tracer(common.ServiceName)
	_, childSpan := tr.Start(ctx, "RepositoryGetUserById")
	defer childSpan.End()
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
	response, err := r.client.GetItem(ctx, input)

	if err != nil {
		r.logger.Error("dynamodb get item failed",
			slog.String("error", err.Error()),
			slog.String("user_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}
	if len(response.Item) == 0 {
		r.logger.Debug("user not found in database",
			slog.String("user_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrUserNotFound
	}
	// unmarshal item to models.user struct
	err = attributevalue.UnmarshalMap(response.Item, &output)
	if err != nil {
		r.logger.Error("error unmarshaling map",
			slog.String("error", err.Error()),
			slog.String("user_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}

	r.logger.Debug("user retrieved from database",
		slog.String("user_id", id),
		slog.String("component", "repository"),
	)

	return output, nil
}

// Function to Create user

func (r *DynamoDBUsersRepository) CreateUser(ctx context.Context, u models.User) (models.UserResponse, error) {
	tr := otel.Tracer(common.ServiceName)
	_, childSpan := tr.Start(ctx, "RepositoryCreateUser")
	defer childSpan.End()

	output := models.UserResponse{}
	u.UserId = r.uuid.New()

	r.logger.Debug("creating user in database",
		slog.String("user_id", u.UserId),
		slog.String("username", u.Username),
		slog.String("email", u.Email),
		slog.String("component", "repository"),
	)

	item, err := attributevalue.MarshalMap(u)
	if err != nil {
		r.logger.Error("error marshaling map",
			slog.String("error", err.Error()),
			slog.String("user_id", u.UserId),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(r.tableUsers),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(userId)"),
	}
	// call dynamodb put item
	_, err = r.client.PutItem(ctx, input)

	if err != nil {
		r.logger.Error("dynamodb error put item",
			slog.String("error", err.Error()),
			slog.String("user_id", u.UserId),
			slog.String("username", u.Username),
			slog.String("component", "repository"),
		)
		return output, common.ErrUserNotCreated
	}

	r.logger.Info("user created in database",
		slog.String("user_id", u.UserId),
		slog.String("username", u.Username),
		slog.String("component", "repository"),
	)

	// create output response
	output.UserId = u.UserId
	output.Username = u.Username
	output.Email = u.Email
	output.Active = u.Active
	return output, nil
}

// Function to Update User

func (r *DynamoDBUsersRepository) UpdateUser(ctx context.Context, u models.User) (models.UserResponse, error) {
	tr := otel.Tracer(common.ServiceName)
	_, childSpan := tr.Start(ctx, "RepositoryUpdateUser")
	defer childSpan.End()

	output := models.UserResponse{}

	item, err := attributevalue.MarshalMap(u)
	if err != nil {
		r.logger.Error("error marshaling map",
			"error", err.Error(),
			"userId", u.UserId,
		)
		return output, common.ErrInternalError
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(r.tableUsers),
		Item:      item,
	}
	// call dynamodb put item
	_, err = r.client.PutItem(ctx, input)

	if err != nil {
		r.logger.Error("dynamodb error put item",
			"error", err.Error(),
			"userId", u.UserId,
		)
		return output, common.ErrUserNotCreated
	}
	// create output response
	output.UserId = u.UserId
	output.Username = u.Username
	output.Active = u.Active

	return output, nil
}

// Function to Get user by email
func (r *DynamoDBUsersRepository) GetUserByEmailorUsername(ctx context.Context, k string, v string) (models.User, error) {
	tr := otel.Tracer(common.ServiceName)
	_, childSpan := tr.Start(ctx, "RepositoryGetUserByEmailorUsername")
	defer childSpan.End()

	output := models.User{}

	// create keycondition dynamodb expression for the query
	keyCondition := expression.Key(k).Equal(expression.Value(v))

	// create expression builder for the keyCondition
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()

	if err != nil {
		r.logger.Error("dynamodb error building expression",
			slog.String("error", err.Error()),
			slog.String("key", k),
			slog.String("value", v),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}

	r.logger.Debug("querying user by email or username",
		slog.String("key", k),
		slog.String("value", v),
		slog.String("component", "repository"),
	)

	// Create the input for the dynamodb query
	queryInput := &dynamodb.QueryInput{
		TableName:                 aws.String(r.tableUsers),
		IndexName:                 aws.String("by_" + k),
		KeyConditionExpression:    expr.KeyCondition(),
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
	}
	// Execute the query

	response, err := r.client.Query(ctx, queryInput)

	if err != nil {
		r.logger.Error("dynamodb error query item",
			slog.String("error", err.Error()),
			slog.String("key", k),
			slog.String("value", v),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}

	if len(response.Items) == 0 {
		r.logger.Debug("user not found in database",
			slog.String("key", k),
			slog.String("value", v),
			slog.String("component", "repository"),
		)
		return output, common.ErrUserNotFound
	}
	//unmarshall dynamodb output
	err = attributevalue.UnmarshalMap(response.Items[0], &output)
	if err != nil {
		r.logger.Error("error unmarshaling map",
			slog.String("error", err.Error()),
			slog.String("key", k),
			slog.String("value", v),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}

	r.logger.Debug("user found in database",
		slog.String("user_id", output.UserId),
		slog.String("key", k),
		slog.String("component", "repository"),
	)

	return output, nil

}
