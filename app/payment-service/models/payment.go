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
	UserId   string `json:"userId" dynamodbav:"userId"` // primary key
	Username string `json:"username" dynamodbav:"username"`
	Email    string `json:"email" dynamodbav:"email"` // global secondary index
	Active   bool   `json:"active" dynamodbav:"active"`
}

type Expense struct {
	ExpenseId string  `json:"expenseId" dynamodbav:"expenseId"` // primary key
	Date      string  `json:"date" dynamodbav:"date"`
	Name      string  `json:"name" dynamodbav:"name"`
	Amount    float64 `json:"amount" dynamodbav:"amount"`
	Status    string  `json:"priority" dynamodbav:"priority"`
}

type PaymentRequest struct {
	UserId    string `json:"userId" dynamodbav:"userId"` // global secondary index
	ExpenseId string `json:"expenseId" dynamodbav:"expenseId"`
}
