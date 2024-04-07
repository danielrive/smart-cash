package models

// Define struct for payments

type Order struct {
	OrderId    string `json:"orderId" dynamodbav:"orderId"` // primary key
	Date       string `json:"date" dynamodbav:"date"`       // global secondary index
	Paid       bool   `json:"paid" dynamodbav:"paid"`
	ExpensesId string `json:"expensesId" dynamodbav:"expensesOd"`
	UserId     string `json:"userId" dynamodbav:"userId"`
}
