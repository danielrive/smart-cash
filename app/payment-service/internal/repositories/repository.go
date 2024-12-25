package repositories

import (
	"context"
	"smart-cash/payment-service/internal/common"

	"log/slog"
	"smart-cash/payment-service/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"go.opentelemetry.io/otel"
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
	tr := otel.Tracer(common.ServiceName)
	_, childSpan := tr.Start(ctx, "RepositoryCreateTransaction")
	defer childSpan.End()

	item, err := attributevalue.MarshalMap(transaction)
	if err != nil {
		r.logger.Error("error while unmarshaling DynamoDB item",
			"error", err.Error(),
			"transactionId", transaction.TransactionId,
		)
		return common.ErrTransactionFailed
	}

	_, err = r.client.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName: aws.String(r.paymentTable),
		Item:      item,
	})

	if err != nil {
		r.logger.Error("dynamodb error while putting item",
			"error", err.Error(),
			"transactionId", transaction.TransactionId,
		)
		return common.ErrTransactionFailed
	}

	return nil

}

// Function to get a user by id

func (r *DynamoDBPaymentRepository) GetTransaction(ctx context.Context, id string) (models.TransactionRequest, error) {
	tr := otel.Tracer(common.ServiceName)
	_, childSpan := tr.Start(ctx, "RepositoryGetTransaction")
	defer childSpan.End()

	output := models.TransactionRequest{}
	// Get bank item by id
	item, err := r.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: aws.String(r.paymentTable),
		Key: map[string]types.AttributeValue{
			"transactionId": &types.AttributeValueMemberS{Value: id},
		},
	})
	if err != nil {
		r.logger.Error("dynamodb couldn't get the item",
			"error", err.Error(),
			"transactionId", id,
		)
		return output, common.ErrInternalError
	}
	if len(item.Item) == 0 {
		r.logger.Info("transaction not found",
			"transactionId", id,
		)
		return output, common.ErrTransactionNotFound
	}

	// Unmarshal the bank item
	err = attributevalue.UnmarshalMap(item.Item, &output)
	if err != nil {
		r.logger.Error("failed to unmarshal attribute value",
			"error", err.Error(),
			"transactionId", id,
		)
		return output, common.ErrInternalError
	}

	return output, nil
}

// Func to update user
func (r *DynamoDBPaymentRepository) UpdateTransaction(ctx context.Context, transaction models.TransactionRequest) error {
	tr := otel.Tracer(common.ServiceName)
	_, childSpan := tr.Start(ctx, "RepositoryUpdateTransaction")
	defer childSpan.End()
	// Marshal the bank item
	update := expression.Set(expression.Name("status"), expression.Value(transaction.Status))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		r.logger.Error("dynamodb update expression couldn't be created",
			"error", err.Error(),
			"transactionId", transaction.TransactionId,
		)
		return common.ErrInternalError
	}
	// Define the key of the item to update
	transactionId, err := attributevalue.Marshal(transaction.TransactionId)
	if err != nil {
		r.logger.Error("dynamodb udpate key couldn't be created",
			"error", err.Error(),
			"transactionId", transaction.TransactionId,
		)
		return common.ErrInternalError
	}

	inputUpdate := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.paymentTable),
		Key:                       map[string]types.AttributeValue{"userId": transactionId},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	}
	_, err = r.client.UpdateItem(context.TODO(), inputUpdate)

	// Update bank item
	if err != nil {
		r.logger.Error("saving could't be updated",
			"error", err.Error(),
			"transactionId", transaction.TransactionId,
		)
		return common.ErrInternalError
	}
	return nil
}
