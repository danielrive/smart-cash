package models

// Define struct for expenses

type Expense struct {
	ExpenseId   string `json:"expenseId" dynamodbav:"expenseId"` // primary key
	Description string `json:"description" dynamodbav:"description"`
	Currency    string `json:"currency" dynamodbav:"currency"`
	Paid        bool   `json:"paid" dynamodbav:"paid"`
	Name        string `json:"name" dynamodbav:"name"`
	Amount      int32  `json:"amount" dynamodbav:"amount"`
}
