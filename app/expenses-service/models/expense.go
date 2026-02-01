package models

import "time"

// Define struct for expenses

type Expense struct {
	ExpenseId   string    `json:"expenseId" dynamodbav:"expenseId" validate:"required"`
	Date        time.Time `json:"date" dynamodbav:"date"`
	Description string    `json:"description" dynamodbav:"description" validate:"required,max=500"`
	Status      string    `json:"status" dynamodbav:"status" validate:"required,oneof=pending paid cancelled"`
	Name        string    `json:"name" dynamodbav:"name" validate:"required,max=100"`
	Amount      float64   `json:"amount" dynamodbav:"amount" validate:"required,gt=0"`
	UserId      string    `json:"userId" dynamodbav:"userId" validate:"required"`
	Category    string    `json:"category" dynamodbav:"category" validate:"required"`
	CreatedAt   time.Time `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt" dynamodbav:"updatedAt"`
	Tags        []string  `json:"tags,omitempty" dynamodbav:"tags,omitempty"`
}

type ExpensesReturn struct {
	ExpenseId   string    `json:"expenseId"`
	Date        time.Time `json:"date"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Status      string    `json:"status"`
	Tags        []string  `json:"tags,omitempty"`
}

type ExpensesPay struct {
	ExpenseId string `json:"expenseId" dynamodbav:"expenseId"` // primary key
}
