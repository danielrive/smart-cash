package models

// Define struct for expenses

type BankUser struct {
	UserId       string        `json:"userId" dynamodbav:"userId"` // Primary key
	Currency     string        `json:"currency" dynamodbav:"currency"`
	Savings      float64       `json:"savings" dynamodbav:"savings"`
	Blocked      bool          `json:"blocked" dynamodbav:"blocked"`
	Transactions []Transaction `json:"Transactions" dynamodbav:"Transactions"`
}

type Transaction struct {
	TransactionId string  `json:"expenseId" dynamodbav:"expenseId"` // primary key
	Date          string  `json:"date" dynamodbav:"date"`
	Amount        float64 `json:"amount" dynamodbav:"amount"`
	Status        string  `json:"status" dynamodbav:"status"`
}

type TransactionRequest struct {
	TransactionId string  `json:"transactionId" dynamodbav:"transactionId"` // primary key
	ExpenseId     string  `json:"expenseId" dynamodbav:"expenseId"`
	Date          string  `json:"date" dynamodbav:"date"`
	Amount        float64 `json:"amount" dynamodbav:"amount"`
	UserId        string  `json:"userId" dynamodbav:"userId"` // global secondary index
	Status        string  `json:"status" dynamodbav:"status"`
}
