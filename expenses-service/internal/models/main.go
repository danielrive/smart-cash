package models

// Define struct for expenses

type Expense struct {
	ExpenseId   string `json:"expenseId" dynamodbav:"expenseId"` // primary key
	Date        string `json:"date" dynamodbav:"date"`
	Description string `json:"description" dynamodbav:"description"`
	Currency    string `json:"currency" dynamodbav:"currency"`
	Paid        bool   `json:"paid" dynamodbav:"paid"`
	Name        string `json:"name" dynamodbav:"name"`
	Amount      int32  `json:"amount" dynamodbav:"amount"`
<<<<<<< HEAD
	Date        string `json:"date" dynamodbav:"date"`         // sort key
	Category    string `json:"category" dynamodbav:"category"` // secondary index
	UserId      string `json:"userId" dynamodbav:"userId"`
=======
	UserId      string `json:"userId" dynamodbav:"userId"` // global secondary index
	Tag         string `json:"tag" dynamodbav:"tag"`       // global secondary index
>>>>>>> 2826218 (update k8 version to 1.29)
}
