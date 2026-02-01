package repositories

import (
	"context"
	"smart-cash/bank-service/internal/common"
	"smart-cash/bank-service/models"
	"smart-cash/utils"
	"smart-cash/utils/logging"

	"log/slog"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type DynamoDBBankRepository struct {
	client    *dynamodb.Client
	logger    *slog.Logger
	bankTable string
}

// loggerWithTrace returns a logger with trace context if available in the context
func (r *DynamoDBBankRepository) loggerWithTrace(ctx context.Context) *slog.Logger {
	return logging.LoggerWithTraceContext(ctx, r.logger)
}

func NewDynamoDBBankRepository(client *dynamodb.Client, bankTable string, logger *slog.Logger) *DynamoDBBankRepository {
	return &DynamoDBBankRepository{
		client:    client,
		bankTable: bankTable,
		logger:    logger,
	}
}

// Function to get a user by id

func (r *DynamoDBBankRepository) GetUser(ctx context.Context, id string) (models.BankUser, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryGetUser", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("GetItem"),
		semconv.DBNameKey.String(r.bankTable),
		attribute.String("db.table", r.bankTable),
		attribute.String("user.id", id),
	)
	defer endSpan()

	output := models.BankUser{}
	// Get bank item by id
	item, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.bankTable),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.loggerWithTrace(ctx).Error("dynamodb couldn't get the item",
			"error", err.Error(),
			"userId", id,
		)
		return output, common.ErrInternalError
	}
	if len(item.Item) == 0 {
		utils.AddSpanEvent(ctx, "bank user not found", attribute.Bool("db.item_found", false))
		utils.SetSpanStatus(ctx, codes.Ok, "bank user not found")
		r.loggerWithTrace(ctx).Debug("bank user not found in database",
			slog.String("user_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrUserNotFound
	}

	// Unmarshal the bank item
	err = attributevalue.UnmarshalMap(item.Item, &output)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.loggerWithTrace(ctx).Error("error unmarshaling bank user item",
			slog.String("error", err.Error()),
			slog.String("user_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}

	utils.AddSpanEvent(ctx, "bank user retrieved successfully",
		attribute.Bool("db.item_found", true),
		attribute.String("user.id", output.UserId),
		attribute.String("user.currency", output.Currency),
		attribute.Float64("user.savings", output.Savings),
	)
	utils.SetSpanStatus(ctx, codes.Ok, "bank user retrieved successfully")

	r.loggerWithTrace(ctx).Debug("bank user retrieved from database",
		slog.String("user_id", id),
		slog.String("component", "repository"),
	)

	return output, nil
}

// Func to update user
func (r *DynamoDBBankRepository) UpdateSavingsUser(ctx context.Context, user models.BankUser) error {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryUpdateSavingsUser", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("UpdateItem"),
		semconv.DBNameKey.String(r.bankTable),
		attribute.String("db.table", r.bankTable),
		attribute.String("user.id", user.UserId),
		attribute.Float64("user.new_savings", user.Savings),
	)
	defer endSpan()

	// Marshal the bank item
	update := expression.Set(expression.Name("savings"), expression.Value(user.Savings))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.loggerWithTrace(ctx).Error("dynamodb update expression couldn't be created",
			"error", err.Error(),
			"userId", user.UserId,
		)
		return common.ErrInternalError
	}
	// Define the key of the item to update
	userId, err := attributevalue.Marshal(user.UserId)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.loggerWithTrace(ctx).Error("dynamodb udpate key couldn't be created",
			"error", err.Error(),
			"userId", user.UserId,
		)
		return common.ErrInternalError
	}

	inputUpdate := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.bankTable),
		Key:                       map[string]types.AttributeValue{"userId": userId},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	}
	_, err = r.client.UpdateItem(ctx, inputUpdate)

	// Update bank item
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.loggerWithTrace(ctx).Error("error updating user savings",
			slog.String("error", err.Error()),
			slog.String("user_id", user.UserId),
			slog.Float64("new_savings", user.Savings),
			slog.String("component", "repository"),
		)
		return common.ErrInternalError
	}

	utils.AddSpanEvent(ctx, "user savings updated successfully",
		attribute.String("user.id", user.UserId),
		attribute.Float64("user.savings", user.Savings),
		attribute.Bool("db.item_updated", true),
	)
	utils.SetSpanStatus(ctx, codes.Ok, "user savings updated successfully")

	r.loggerWithTrace(ctx).Info("user savings updated in database",
		slog.String("user_id", user.UserId),
		slog.Float64("new_savings", user.Savings),
		slog.String("component", "repository"),
	)

	return nil
}
