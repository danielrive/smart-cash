package models

// Define struct for expenses

type BankUser struct {
	UserId    string  `json:"userId" dynamodbav:"userId"` // Primary key
	FirstName string  `json:"name" dynamodbav:"firstName"`
	LastName  string  `json:"lastName" dynamodbav:"lastName"`
	Currency  string  `json:"currency" dynamodbav:"currency"`
	Savings   float64 `json:"savings" dynamodbav:"savings"`
	Blocked   bool    `json:"blocked" dynamodbav:"blocked"`
}
