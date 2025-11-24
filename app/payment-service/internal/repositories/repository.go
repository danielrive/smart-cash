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

// Create Payment

func (r *DynamoDBPaymentRepository) CreatePayment(ctx context.Context, payment models.PaymentRequest) (models.PaymentResponse, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryCreatePayment", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("PutItem"),
		semconv.DBNameKey.String(r.paymentTable),
		attribute.String("db.table", r.paymentTable),
		attribute.String("payment.id", payment.PaymentId),
		attribute.String("expense.id", payment.ExpenseId),
		attribute.String("user.id", payment.UserId),
	)
	defer endSpan()

	r.logger.Debug("creating payment in database",
		slog.String("transaction_id", payment.PaymentId),
		slog.String("expense_id", payment.ExpenseId),
		slog.String("component", "repository"),
	)

	item, err := attributevalue.MarshalMap(payment)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("error marshaling payment item",
			slog.String("error", err.Error()),
			slog.String("transaction_id", payment.PaymentId),
			slog.String("component", "repository"),
		)
		return models.PaymentResponse{}, common.ErrPaymentFailed
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: aws.String(r.paymentTable),
		Item:      item,
	})

	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("dynamodb error putting payment item",
			slog.String("error", err.Error()),
			slog.String("transaction_id", payment.PaymentId),
			slog.String("component", "repository"),
		)
		return models.PaymentResponse{}, common.ErrPaymentFailed
	}

	utils.AddSpanEvent(ctx, "payment created successfully",
		attribute.String("payment.id", payment.PaymentId),
		attribute.Bool("db.item_created", true),
	)
	utils.SetSpanStatus(ctx, codes.Ok, "payment created successfully")

	r.logger.Info("payment created in database",
		slog.String("transaction_id", payment.PaymentId),
		slog.String("component", "repository"),
	)

	return createPaymentReturn(payment), nil

}

// Function to get a user by id

func (r *DynamoDBPaymentRepository) GetPayment(ctx context.Context, id string) (models.PaymentRequest, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryGetPayment", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("GetItem"),
		semconv.DBNameKey.String(r.paymentTable),
		attribute.String("db.table", r.paymentTable),
		attribute.String("payment.id", id),
	)
	defer endSpan()

	output := models.PaymentRequest{}
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
		utils.AddSpanEvent(ctx, "payment not found", attribute.Bool("db.item_found", false))
		utils.SetSpanStatus(ctx, codes.Ok, "payment not found")
		r.logger.Debug("payment not found in database",
			slog.String("transaction_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrPaymentNotFound
	}

	// Unmarshal the payment item
	err = attributevalue.UnmarshalMap(item.Item, &output)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("error unmarshaling payment item",
			slog.String("error", err.Error()),
			slog.String("transaction_id", id),
			slog.String("component", "repository"),
		)
		return output, common.ErrInternalError
	}

	utils.AddSpanEvent(ctx, "payment retrieved successfully",
		attribute.Bool("db.item_found", true),
		attribute.String("payment.id", output.PaymentId),
		attribute.String("payment.status", output.Status),
	)
	utils.SetSpanStatus(ctx, codes.Ok, "payment retrieved successfully")

	r.logger.Debug("payment retrieved from database",
		slog.String("transaction_id", id),
		slog.String("component", "repository"),
	)

	return output, nil
}

// Func to update user
func (r *DynamoDBPaymentRepository) UpdatePayment(ctx context.Context, payment models.PaymentRequest) error {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "RepositoryUpdatePayment", "repository",
		semconv.DBSystemKey.String("dynamodb"),
		semconv.DBOperationKey.String("UpdateItem"),
		semconv.DBNameKey.String(r.paymentTable),
		attribute.String("db.table", r.paymentTable),
		attribute.String("payment.id", payment.PaymentId),
	)
	defer endSpan()

	// Marshal the bank item
	update := expression.Set(expression.Name("status"), expression.Value(payment.Status))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("dynamodb update expression couldn't be created",
			"error", err.Error(),
			"transactionId", payment.PaymentId,
		)
		return common.ErrInternalError
	}
	// Define the key of the item to update
	transactionId, err := attributevalue.Marshal(payment.PaymentId)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		r.logger.Error("dynamodb udpate key couldn't be created",
			"error", err.Error(),
			"transactionId", payment.PaymentId,
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
			"transactionId", payment.PaymentId,
		)
		return common.ErrInternalError
	}

	utils.AddSpanEvent(ctx, "payment updated successfully",
		attribute.String("payment.status", payment.Status),
		attribute.Bool("db.item_updated", true),
	)
	utils.SetSpanStatus(ctx, codes.Ok, "payment updated successfully")

	return nil
}

func createPaymentReturn(payment models.PaymentRequest) models.PaymentResponse {
	return models.PaymentResponse{
		PaymentId: payment.PaymentId,
		ExpenseId: payment.ExpenseId,
		Date:      payment.Date,
		Status:    payment.Status,
	}
}
