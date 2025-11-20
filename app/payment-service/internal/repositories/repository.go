package repositories

import (
	"context"
	"smart-cash/payment-service/internal/common"
	"smart-cash/payment-service/models"
	"smart-cash/utils"

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

type DynamoDBPaymentRepository struct {
	client       *dynamodb.Client
	logger       *slog.Logger
	paymentTable string
}

func NewDynamoDBPaymentRepository(client *dynamodb.Client, paymentTable string, logger *slog.Logger) *DynamoDBPaymentRepository {
	return &DynamoDBPaymentRepository{
		client:       client,
		paymentTable: paymentTable,
		logger:       logger,
	}
}

// Create Transaction

func (r *DynamoDBPaymentRepository) CreateTransaction(ctx context.Context, transaction models.TransactionRequest) error {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryCreateTransaction", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("PutItem"),
		semconv.DBNameKey.String(r.paymentTable),
		attribute.String("db.table", r.paymentTable),
		attribute.String("transaction.id", transaction.TransactionId),
		attribute.String("expense.id", transaction.ExpenseId),
		attribute.String("user.id", transaction.UserId),
	)
	defer endSpan()

	r.logger.Debug("creating transaction in database",
		slog.String("transaction_id", transaction.TransactionId),
		slog.String("expense_id", transaction.ExpenseId),
		slog.String("component", "repository"),
	)

	item, err := attributevalue.MarshalMap(transaction)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("error marshaling transaction item",
			slog.String("error", err.Error()),
			slog.String("transaction_id", transaction.TransactionId),
			slog.String("component", "repository"),
		)
		return common.ErrTransactionFailed
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.paymentTable),
		Item:      item,
	})

	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("dynamodb error putting transaction item",
			slog.String("error", err.Error()),
			slog.String("transaction_id", transaction.TransactionId),
			slog.String("component", "repository"),
		)
		return common.ErrTransactionFailed
	}

	utils.AddSpanEvent(ctx, "transaction created successfully",
		attribute.String("transaction.id", transaction.TransactionId),
		attribute.Bool("db.item_created", true),
	)
	utils.SetSpanStatus(ctx, codes.Ok, "transaction created successfully")

	r.logger.Info("transaction created in database",
		slog.String("transaction_id", transaction.TransactionId),
		slog.String("component", "repository"),
	)

	return nil

}

// Function to get a user by id

func (r *DynamoDBPaymentRepository) GetTransaction(ctx context.Context, id string) (models.TransactionRequest, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryGetTransaction", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("GetItem"),
		semconv.DBNameKey.String(r.paymentTable),
		attribute.String("db.table", r.paymentTable),
		attribute.String("transaction.id", id),
	)
	defer endSpan()

	output := models.TransactionRequest{}
	// Get bank item by id
	item, err := r.client.GetItem(ctx, &dynamodb.GetItemInput{
		TableName: aws.String(r.paymentTable),
		Key: map[string]types.AttributeValue{
			"transactionId": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("dynamodb couldn't get the item",
			"error", err.Error(),
			"transactionId", id,
		)
		return output, common.ErrInternalError
	}
	if len(item.Item) == 0 {
		utils.AddSpanEvent(ctx, "transaction not found", attribute.Bool("db.item_found", false))
		utils.SetSpanStatus(ctx, codes.Ok, "transaction not found")
		r.logger.Debug("transaction not found in database",
			slog.String("transaction_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrTransactionNotFound
	}

	// Unmarshal the transaction item
	err = attributevalue.UnmarshalMap(item.Item, &output)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("error unmarshaling transaction item",
			slog.String("error", err.Error()),
			slog.String("transaction_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}

	utils.AddSpanEvent(ctx, "transaction retrieved successfully",
		attribute.Bool("db.item_found", true),
		attribute.String("transaction.id", output.TransactionId),
		attribute.String("transaction.status", output.Status),
	)
	utils.SetSpanStatus(ctx, codes.Ok, "transaction retrieved successfully")

	r.logger.Debug("transaction retrieved from database",
		slog.String("transaction_id", id),
		slog.String("component", "repository"),
	)

	return output, nil
}

// Func to update user
func (r *DynamoDBPaymentRepository) UpdateTransaction(ctx context.Context, transaction models.TransactionRequest) error {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryUpdateTransaction", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("UpdateItem"),
		semconv.DBNameKey.String(r.paymentTable),
		attribute.String("db.table", r.paymentTable),
		attribute.String("transaction.id", transaction.TransactionId),
	)
	defer endSpan()

	// Marshal the bank item
	update := expression.Set(expression.Name("status"), expression.Value(transaction.Status))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("dynamodb update expression couldn't be created",
			"error", err.Error(),
			"transactionId", transaction.TransactionId,
		)
		return common.ErrInternalError
	}
	// Define the key of the item to update
	transactionId, err := attributevalue.Marshal(transaction.TransactionId)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("dynamodb udpate key couldn't be created",
			"error", err.Error(),
			"transactionId", transaction.TransactionId,
		)
		return common.ErrInternalError
	}

	inputUpdate := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.paymentTable),
		Key:                       map[string]types.AttributeValue{"transactionId": transactionId},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	}
	_, err = r.client.UpdateItem(ctx, inputUpdate)

	// Update bank item
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("saving could't be updated",
			"error", err.Error(),
			"transactionId", transaction.TransactionId,
		)
		return common.ErrInternalError
	}

	utils.AddSpanEvent(ctx, "transaction updated successfully",
		attribute.String("transaction.status", transaction.Status),
		attribute.Bool("db.item_updated", true),
	)
	utils.SetSpanStatus(ctx, codes.Ok, "transaction updated successfully")

	return nil
}
