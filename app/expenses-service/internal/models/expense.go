package models

// Define struct for expenses

type Expense struct {
	ExpenseId    string  `json:"expenseId" dynamodbav:"expenseId"` // primary key
	Date         string  `json:"date" dynamodbav:"date"`
	Description  string  `json:"description" dynamodbav:"description"`
	Currency     string  `json:"currency" dynamodbav:"currency"`
	Paid         bool    `json:"paid" dynamodbav:"paid"`
	Name         string  `json:"name" dynamodbav:"name"`
	Amount       float64 `json:"amount" dynamodbav:"amount"`
	UserId       string  `json:"userId" dynamodbav:"userId"`     // global secondary index
	Category     string  `json:"category" dynamodbav:"category"` // global secondary index
	AutomaticPay bool    `json:"automaticPay" dynamodbav:"automaticPay"`
	Priority     int     `json:"priority" dynamodbav:"priority"`
	ScheduledTo  string  `json:"scheduledTo" dynamodbav:"scheduledTo"`
}

type ExpensesReturn struct {
	ExpenseId string `json:"expenseId" dynamodbav:"expenseId"` // primary key
	Date      string `json:"date" dynamodbav:"date"`
	Name      string `json:"name" dynamodbav:"name"`
}
