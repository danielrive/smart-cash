package models

type TransactionRequest struct {
	TransactionId string  `json:"transactionId" dynamodbav:"transactionId"` // primary key
	ExpenseId     string  `json:"expenseId" dynamodbav:"expenseId"`
	Date          string  `json:"date" dynamodbav:"date"`
	Amount        float64 `json:"amount" dynamodbav:"amount"`
	UserId        string  `json:"userId" dynamodbav:"userId"` // global secondary index
	Status        string  `json:"status" dynamodbav:"status"`
}

type User struct {
	UserId string `json:"userId" dynamodbav:"userId"` // global secondary index
	Status string `json:"status" dynamodbav:"status"`
}

type Expense struct {
	ExpenseId string  `json:"expenseId" dynamodbav:"expenseId"`
	UserId    string  `json:"userId" dynamodbav:"userId"` // global secondary index
	Amount    float64 `json:"amount" dynamodbav:"amount"`
}

type PaymentRequest struct {
	UserId    string `json:"userId" dynamodbav:"userId"` // global secondary index
	ExpenseId string `json:"expenseId" dynamodbav:"expenseId"`
}
