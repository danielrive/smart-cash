package models

// Define struct for expenses

type Expense struct {
	ExpenseId   string `json:"expenseId" dynamodbav:"expenseId"` // primary key
	Description string `json:"description" dynamodbav:"description"`
	Currency    string `json:"currency" dynamodbav:"currency"`
	Paid        bool   `json:"paid" dynamodbav:"paid"`
	Name        string `json:"name" dynamodbav:"name"`
	Amount      int32  `json:"amount" dynamodbav:"amount"`
	Date        string `json:"date" dynamodbav:"date"`         // sort key
	Category    string `json:"category" dynamodbav:"category"` // secondary index
	UserId      string `json:"userId" dynamodbav:"userId"`
}
