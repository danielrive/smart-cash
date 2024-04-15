package repositories

import (
	"context"
	"log"
	"smart-cash/payment-service/internal/common"
	"smart-cash/payment-service/internal/models"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// define uuid interface
type UUIDHelper interface {
	New() string
}

// Define DynamoDB repository struct

type DynamoDBPaymentRepository struct {
	client       *dynamodb.Client
	paymentTable string
	uuid         UUIDHelper
}

func NewDynamoDBPaymentRepository(client *dynamodb.Client, paymentTable string, uuid UUIDHelper) *DynamoDBPaymentRepository {
	return &DynamoDBPaymentRepository{
		client:       client,
		paymentTable: paymentTable,
		uuid:         uuid,
	}
}

// Create a order
func (c *DynamoDBPaymentRepository) CreateOrder(order models.Order) (string, error) {

	order.OrderId = c.uuid.New()
	item, err := attributevalue.MarshalMap(order)
	if err != nil {
		log.Println(err)
		return "", err
	}
	input := &dynamodb.PutItemInput{
		TableName:           aws.String(c.paymentTable),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(expenseId)"),
	}
	// call dynamodb put item
	_, err = c.client.PutItem(context.TODO(), input)

	if err != nil {
		log.Println(common.ErrOrderNoCreated, err)
		return "", common.ErrOrderNoCreated
	}
	return order.OrderId, nil

}

// Get order by id
func (c *DynamoDBPaymentRepository) GetOrderById(orderId string) (models.Order, error) {
	output := models.Order{}

	// create input for get item
	input := &dynamodb.GetItemInput{
		TableName: aws.String(c.paymentTable),
		Key: map[string]types.AttributeValue{
			"orderId": &types.AttributeValueMemberS{
				Value: orderId,
			},
		},
	}
	// call getItem
	response, err := c.client.GetItem(context.TODO(), input)

	if err != nil {
		log.Println("error", err)
		return output, err
	}
	if len(response.Item) == 0 {
		log.Println("error", common.ErrOrderNotFound, err)
		return output, common.ErrOrderNotFound
	}
	// unmarshal item to models.user struct
	err = attributevalue.UnmarshalMap(response.Item, &output)
	if err != nil {
		log.Println("error", err)
		return output, err
	}
	return output, nil
}
