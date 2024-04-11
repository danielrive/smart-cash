package repositories

import (
	"context"
	"log"
	"smart-cash/bank-service/internal/common"
	"smart-cash/bank-service/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// Define DynamoDB repository struct

type DynamoDBTransactionRepository struct {
	client            *dynamodb.Client
	transactionsTable string
}

func NewDynamoDBTransactionRepository(client *dynamodb.Client, transactionsTable string) *DynamoDBTransactionRepository {
	return &DynamoDBTransactionRepository{
		client:            client,
		transactionsTable: transactionsTable,
	}
}

// repository for bank transactions

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

////// Transactions methods //////
// Create a new transaction

func (c *DynamoDBTransactionRepository) CreateTransaction(transaction models.Transaction) error {
	item, err := attributevalue.MarshalMap(transaction)
	if err != nil {
		log.Println(err)
		return err
	}
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(c.transactionsTable),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(transactionId)"),
	}
	// call dynamodb put item
	_, err = c.client.PutItem(context.TODO(), input)

	if err != nil {
		log.Println(common.ErrTransactionNoCreated, err)
		return common.ErrTransactionNoCreated
	}
	return nil
}

// get transaction by ID

func (c *DynamoDBTransactionRepository) GetTransactionById(transactionId string) (models.Transaction, error) {
	output := models.Transaction{}
	input := &dynamodb.GetItemInput{
		TableName: aws.String(c.transactionsTable),
		Key: map[string]types.AttributeValue{
			"transactionId": &types.AttributeValueMemberS{
				Value: transactionId,
			},
		},
	}
	result, err := c.client.GetItem(context.TODO(), input)
	if err != nil {
		log.Println(common.ErrTransactionNotFound, err)
		return output, common.ErrTransactionNotFound
	}

	if len(result.Item) == 0 {
		log.Println(common.ErrTransactionNotFound, err)
		return output, common.ErrTransactionNotFound
	}
	// unmarshal item to models.user struct
	err2 := attributevalue.UnmarshalMap(result.Item, &output)
	if err2 != nil {
		log.Println(err2)
		return output, err2
	}
	return output, nil
}

////// Bank methods //////

// get user info by id

func (c *DynamoDBBankRepository) GetInfoById(userId string) (models.User, error) {
	output := models.User{}
	input := &dynamodb.GetItemInput{
		TableName: aws.String(c.bankTable),
		Key: map[string]types.AttributeValue{
			"userId": &types.AttributeValueMemberS{
				Value: userId,
			},
		},
	}
	result, err := c.client.GetItem(context.TODO(), input)
	if err != nil {
		log.Println(common.ErrUserNotFound, err)
		return output, common.ErrUserNotFound
	}

	if len(result.Item) == 0 {
		log.Println(common.ErrUserNotFound, err)
		return output, common.ErrUserNotFound
	}
	// unmarshal item to models.user struct
	err2 := attributevalue.UnmarshalMap(result.Item, &output)
	if err2 != nil {
		log.Println(err2)
		return output, err2
	}
	return output, nil
}

// update user info

func (c *DynamoDBBankRepository) UpdateInfo(user models.User) error {
	item, err := attributevalue.MarshalMap(user)
	if err != nil {
		log.Println(err)
		return err
	}
	input := &dynamodb.PutItemInput{
		TableName: aws.String(c.bankTable),
		Item:      item,
	}
	// call dynamodb put item
	_, err = c.client.PutItem(context.TODO(), input)

	if err != nil {
		log.Println(common.ErrUserNoUpdated, err)
		return common.ErrUserNoUpdated
	}
	return nil
}
