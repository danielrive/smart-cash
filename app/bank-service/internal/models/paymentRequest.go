package models

type PaymentRequest struct {
	ExpenseId string  `json:"expenseId" dynamodbav:"expenseId"` // primary key
	Date      string  `json:"date" dynamodbav:"date"`
	Amount    float64 `json:"amount" dynamodbav:"amount"`
	UserId    string  `json:"userId" dynamodbav:"userId"` // global secondary index
	Status    string  `json:"status" dynamodbav:"status"`
}
