package models

// Define struct for expenses

type Transaction struct {
	TransactionId string  `json:"transactionId" dynamodbav:"transactionId"`
	OrderId       string  `json:"orderId" dynamodbav:"orderId"`
	Date          string  `json:"date" dynamodbav:"date"`
	Currency      string  `json:"currency" dynamodbav:"currency"`
	UserId        string  `json:"userId" dynamodbav:"userId"`
	Amount        float32 `json:"amount" dynamodbav:"amount"`
}

type User struct {
	UserID  string  `json:"userId" dynamodbav:"userId"`
	Name    string  `json:"name" dynamodbav:"name"`
	Balance float64 `json:"balance" dynamodbav:"balance"`
}
