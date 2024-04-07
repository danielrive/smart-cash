package models

// Define struct for expenses

type Transaction struct {
	TransactionID string  `json:"transactionID" dynamodbav:"transactionID"`
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
