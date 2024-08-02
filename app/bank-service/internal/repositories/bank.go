package repositories

import (
	"context"
	"smart-cash/bank-service/internal/common"

	"log"
	"smart-cash/bank-service/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/expression"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type DynamoDBBankRepository struct {
	client    *dynamodb.Client
	bankTable string
}

func NewDynamoDBBankRepository(client *dynamodb.Client, bankTable string) *DynamoDBBankRepository {
	return &DynamoDBBankRepository{
		client:    client,
		bankTable: bankTable,
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
		log.Printf("Internal error while getting item from DynamoDB: %v", err)
		return output, common.ErrInternalError
	}
	if len(item.Item) == 0 {
		log.Printf("User with id %s not found in DynamoDB", id)
		return output, common.ErrTransactionFailed
	}

	// Unmarshal the bank item
	err = attributevalue.UnmarshalMap(item.Item, &output)
	if err != nil {
		log.Printf("Internal error while unmarshaling DynamoDB item: %v", err)
		return output, common.ErrInternalError
	}

	return output, nil
}

// Func to update user

func (r *DynamoDBBankRepository) UpdateSavingsUser(user models.BankUser) (models.BankUser, error) {
	// Marshal the bank item
	update := expression.Set(expression.Name("savings"), expression.Value(user.Savings))
	expr, err := expression.NewBuilder().WithUpdate(update).Build()
	if err != nil {
		log.Printf("DynamoDB udpate expression couldn't be created: %v", err)
		return models.BankUser{}, common.ErrInternalError
	}
	// Define the key of the item to update
	userId, err := attributevalue.Marshal(user.UserId)
	if err != nil {
		log.Printf("DynamoDB udpate key couldn't be created: %v", err)
		return models.BankUser{}, common.ErrInternalError
	}

	inputUpdate := &dynamodb.UpdateItemInput{
		TableName:                 aws.String(r.bankTable),
		Key:                       map[string]types.AttributeValue{"userId": userId},
		ExpressionAttributeNames:  expr.Names(),
		ExpressionAttributeValues: expr.Values(),
		UpdateExpression:          expr.Update(),
		ReturnValues:              types.ReturnValueUpdatedNew,
	}
	response, err := r.client.UpdateItem(context.TODO(), inputUpdate)

	// Update bank item
	if err != nil {
		log.Printf("Saving for user %v couldn't be updated:", err)
		return models.BankUser{}, common.ErrInternalError
	}
	// Unmarshal the updated bank item
	// Unmarshal the bank item
	log.Println(response)
	return models.BankUser{}, nil
}
