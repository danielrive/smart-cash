package repositories

import (
	"context"
	"smart-cash/bank-service/internal/common"

	"log/slog"
	"smart-cash/bank-service/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
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

func (r *DynamoDBBankRepository) GetUser(id string) (models.BankUser, error) {
	output := models.BankUser{}
	// Get bank item by id
	item, err := r.client.GetItem(context.TODO(), &dynamodb.GetItemInput{
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
		r.logger.Info("user not found",
			"userId", id,
		)
		return output, common.ErrUserNotFound
	}

	// Unmarshal the bank item
	err = attributevalue.UnmarshalMap(item.Item, &output)
	if err != nil {
		r.logger.Error("failed to unmarshal attribute value",
			"error", err.Error(),
			"userId", id,
		)
		return output, common.ErrInternalError
	}

	return output, nil
}

// Func to update user
func (r *DynamoDBBankRepository) UpdateSavingsUser(user models.BankUser) error {
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
	_, err = r.client.UpdateItem(context.TODO(), inputUpdate)

	// Update bank item
	if err != nil {
		r.logger.Error("saving could't be updated",
			"error", err.Error(),
			"userId", user.UserId,
		)
		return common.ErrInternalError
	}
	return nil
}
