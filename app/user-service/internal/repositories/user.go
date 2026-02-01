package repositories

import (
	"context"
	"errors"
	"log/slog"
	"smart-cash/user-service/internal/common"
	"smart-cash/user-service/models"
	"smart-cash/utils"
	"smart-cash/utils/logging"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
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

// loggerWithTrace returns a logger with trace context if available in the context
func (r *DynamoDBUsersRepository) loggerWithTrace(ctx context.Context) *slog.Logger {
	return logging.LoggerWithTraceContext(ctx, r.logger)
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
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryGetUserById",
		"repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("GetItem"),
		semconv.DBNameKey.String(r.tableUsers),
		attribute.String("user.id", id),
		attribute.String("db.table", r.tableUsers),
	)

	defer endSpan()

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
		utils.RecordSpanError(ctx, err)

		r.loggerWithTrace(ctx).Error("dynamodb get item failed",
			slog.String("error", err.Error()),
			slog.String("user_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}
	if len(response.Item) == 0 {
		utils.AddSpanEvent(ctx, "user not found", attribute.Bool("db.item_found", false))

		r.loggerWithTrace(ctx).Debug("user not found in database",
			slog.String("user_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrUserNotFound
	}

	// unmarshal item to models.user struct
	err = attributevalue.UnmarshalMap(response.Item, &output)
	if err != nil {
		utils.RecordSpanError(ctx, err)

		r.loggerWithTrace(ctx).Error("error unmarshaling map",
			slog.String("error", err.Error()),
			slog.String("user_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}

	// Success - mark span as successful and add result info
	utils.AddSpanEvent(ctx, "user retrieved successfully",
		attribute.Bool("db.item_found", true),
		attribute.String("user.id", output.UserId),
	)
	utils.SetSpanStatus(ctx, codes.Ok, "user retrieved successfully")

	r.loggerWithTrace(ctx).Debug("user retrieved from database",
		slog.String("user_id", id),
		slog.String("component", "repository"),
	)

	return output, nil
}

// Function to Create user

func (r *DynamoDBUsersRepository) CreateUser(ctx context.Context, u models.User) (models.UserResponse, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryCreateUser", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("PutItem"),
		semconv.DBNameKey.String(r.tableUsers),
		attribute.String("db.table", r.tableUsers),
		attribute.String("user.username", u.Username),
		attribute.String("user.email", u.Email),
	)
	defer endSpan()

	output := models.UserResponse{}
	u.UserId = r.uuid.New()

	// Add the generated user ID to the span
	utils.AddSpanEvent(ctx, "user.id.generated", attribute.String("user.id", u.UserId))

	r.loggerWithTrace(ctx).Debug("creating user in database",
		slog.String("user_id", u.UserId),
		slog.String("username", u.Username),
		slog.String("email", u.Email),
		slog.String("component", "repository"),
	)

	userItem, err := attributevalue.MarshalMap(u)
	if err != nil {
		utils.RecordSpanError(ctx, err)

		r.loggerWithTrace(ctx).Error("error marshaling map",
			slog.String("error", err.Error()),
			slog.String("user_id", u.UserId),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}

	input := &dynamodb.TransactWriteItemsInput{
		TransactItems: []types.TransactWriteItem{
			{
				// 1. The Actual User Item
				Put: &types.Put{
					TableName:           aws.String(r.tableUsers),
					Item:                userItem,
					ConditionExpression: aws.String("attribute_not_exists(userId)"),
				},
			},
			{
				// 2. The Username Marker (prevents duplicate usernames)
				Put: &types.Put{
					TableName: aws.String(r.tableUsers),
					Item: map[string]types.AttributeValue{
						"userId": &types.AttributeValueMemberS{Value: "USERNAME#" + u.Username},
					},
					ConditionExpression: aws.String("attribute_not_exists(userId)"),
				},
			},
			{
				// 3. The Email Marker (prevents duplicate emails)
				Put: &types.Put{
					TableName: aws.String(r.tableUsers),
					Item: map[string]types.AttributeValue{
						"userId": &types.AttributeValueMemberS{Value: "EMAIL#" + u.Email},
					},
					ConditionExpression: aws.String("attribute_not_exists(userId)"),
				},
			},
		},
	}

	_, err = r.client.TransactWriteItems(ctx, input)

	if err != nil {
		var tcf *types.TransactionCanceledException
		if errors.As(err, &tcf) {
			for _, reason := range tcf.CancellationReasons {
				if *reason.Code == "ConditionalCheckFailed" {
					r.loggerWithTrace(ctx).Warn("user already exists (username or email)", slog.String("username", u.Username))
					// IMPORTANT: Return a specific conflict error so your API returns 409
					return models.UserResponse{}, common.ErrUserAlreadyExists
				}
			}
		}
		return models.UserResponse{}, common.ErrUserNotCreated
	}

	if err != nil {
		utils.RecordSpanError(ctx, err)

		r.loggerWithTrace(ctx).Error("dynamodb error put item",
			slog.String("error", err.Error()),
			slog.String("user_id", u.UserId),
			slog.String("username", u.Username),
			slog.String("component", "repository"),
		)
		return output, common.ErrUserNotCreated
	}

	// Success - mark span as successful
	utils.AddSpanEvent(ctx, "user created successfully",
		attribute.String("user.id", u.UserId),
		attribute.Bool("db.item_created", true),
	)
	utils.SetSpanStatus(ctx, codes.Ok, "user created successfully")

	r.loggerWithTrace(ctx).Info("user created in database",
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
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryUpdateUser", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("PutItem"),
		semconv.DBNameKey.String(r.tableUsers),
		attribute.String("db.table", r.tableUsers),
		attribute.String("user.id", u.UserId),
	)
	defer endSpan()

	output := models.UserResponse{}

	item, err := attributevalue.MarshalMap(u)
	if err != nil {
		utils.RecordSpanError(ctx, err)

		r.loggerWithTrace(ctx).Error("error marshaling map",
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
		utils.RecordSpanError(ctx, err)

		r.loggerWithTrace(ctx).Error("dynamodb error put item",
			"error", err.Error(),
			"userId", u.UserId,
		)
		return output, common.ErrUserNotCreated
	}

	// Success
	utils.AddSpanEvent(ctx, "user updated successfully", attribute.Bool("db.item_updated", true))
	utils.SetSpanStatus(ctx, codes.Ok, "user updated successfully")

	// create output response
	output.UserId = u.UserId
	output.Username = u.Username
	output.Active = u.Active

	return output, nil
}

// Function to Get user by email
func (r *DynamoDBUsersRepository) GetUserByEmailorUsername(ctx context.Context, k string, v string) (models.User, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryGetUserByEmailorUsername", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("Query"),
		semconv.DBNameKey.String(r.tableUsers),
		attribute.String("db.table", r.tableUsers),
		attribute.String("db.index", "by_"+k),
		attribute.String("query.key", k),
		attribute.String("query.value", v),
	)
	defer endSpan()

	output := models.User{}

	// create keycondition dynamodb expression for the query
	keyCondition := expression.Key(k).Equal(expression.Value(v))

	// create expression builder for the keyCondition
	expr, err := expression.NewBuilder().WithKeyCondition(keyCondition).Build()

	if err != nil {
		utils.RecordSpanError(ctx, err)

		r.loggerWithTrace(ctx).Error("dynamodb error building expression",
			slog.String("error", err.Error()),
			slog.String("key", k),
			slog.String("value", v),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}

	r.loggerWithTrace(ctx).Debug("querying user by email or username",
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
		utils.RecordSpanError(ctx, err)

		r.loggerWithTrace(ctx).Error("dynamodb error query item",
			slog.String("error", err.Error()),
			slog.String("key", k),
			slog.String("value", v),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}

	if len(response.Items) == 0 {
		utils.AddSpanEvent(ctx, "user not found", attribute.Bool("db.item_found", false))
		utils.SetSpanStatus(ctx, codes.Ok, "user not found")

		r.loggerWithTrace(ctx).Debug("user not found in database",
			slog.String("key", k),
			slog.String("value", v),
			slog.String("component", "repository"),
		)
		return output, common.ErrUserNotFound
	}

	//unmarshall dynamodb output
	err = attributevalue.UnmarshalMap(response.Items[0], &output)
	if err != nil {
		utils.RecordSpanError(ctx, err)

		r.loggerWithTrace(ctx).Error("error unmarshaling map",
			slog.String("error", err.Error()),
			slog.String("key", k),
			slog.String("value", v),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}

	// Success
	utils.AddSpanEvent(ctx, "user found successfully",
		attribute.Bool("db.item_found", true),
		attribute.String("user.id", output.UserId),
	)
	utils.SetSpanStatus(ctx, codes.Ok, "user found successfully")

	r.loggerWithTrace(ctx).Debug("user found in database",
		slog.String("user_id", output.UserId),
		slog.String("key", k),
		slog.String("component", "repository"),
	)

	return output, nil

}
