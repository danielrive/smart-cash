package repositories

import (
	"context"
	"smart-cash/bank-service/internal/common"

	"log/slog"
	"smart-cash/bank-service/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.opentelemetry.io/otel"
)

type DynamoDBBankRepository struct {
	client    *dynamodb.Client
	logger    *slog.Logger
	bankTable string
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
	tr := otel.Tracer(common.ServiceName)
	_, childSpan := tr.Start(ctx, "RepositoryGetUser")
	defer childSpan.End()

	output := models.BankUser{}
	// Get bank item by id
	item, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.bankTable),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		r.logger.Error("dynamodb couldn't get the item",
			"error", err.Error(),
			"userId", id,
		)
		return output, common.ErrInternalError
	}
	if len(item.Item) == 0 {
		r.logger.Debug("bank user not found in database",
			slog.String("user_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrUserNotFound
	}

	// Unmarshal the bank item
	err = attributevalue.UnmarshalMap(item.Item, &output)
	if err != nil {
		r.logger.Error("error unmarshaling bank user item",
			slog.String("error", err.Error()),
			slog.String("user_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}

	r.logger.Debug("bank user retrieved from database",
		slog.String("user_id", id),
		slog.String("component", "repository"),
	)

	return output, nil
}

// Func to update user
func (r *DynamoDBBankRepository) UpdateSavingsUser(ctx context.Context, user models.BankUser) error {
	tr := otel.Tracer(common.ServiceName)
	_, childSpan := tr.Start(ctx, "RepositoryUpdateSavingsUser")
	defer childSpan.End()
	// Marshal the bank item
	update := expression.Set(expression.Name("savings"), expression.Value(user.Savings))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		r.logger.Error("dynamodb update expression couldn't be created",
			"error", err.Error(),
			"userId", user.UserId,
		)
		return common.ErrInternalError
	}
	// Define the key of the item to update
	userId, err := attributevalue.Marshal(user.UserId)
	if err != nil {
		r.logger.Error("dynamodb udpate key couldn't be created",
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
		r.logger.Error("error updating user savings",
			slog.String("error", err.Error()),
			slog.String("user_id", user.UserId),
			slog.Float64("new_savings", user.Savings),
			slog.String("component", "repository"),
		)
		return common.ErrInternalError
	}

	r.logger.Info("user savings updated in database",
		slog.String("user_id", user.UserId),
		slog.Float64("new_savings", user.Savings),
		slog.String("component", "repository"),
	)

	return nil
}
