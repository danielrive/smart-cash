package models

import "time"

// PaymentStatus constants
const (
	PaymentStatusPending    = "pending"
	PaymentStatusProcessing = "processing"
	PaymentStatusCompleted  = "completed"
	PaymentStatusFailed     = "failed"
	PaymentStatusCancelled  = "cancelled"
)

type PaymentRequest struct {
	PaymentId string    `json:"paymentId" dynamodbav:"paymentId"` // primary key
	UserId    string    `json:"userId" dynamodbav:"userId"`       // global secondary index
	ExpenseId string    `json:"expenseId" dynamodbav:"expenseId"`
	Amount    float64   `json:"amount" dynamodbav:"amount"`
	Date      time.Time `json:"date" dynamodbav:"date"`
	Status    string    `json:"status" dynamodbav:"status"`
	CreatedAt time.Time `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" dynamodbav:"updatedAt"`
}

type User struct {
	UserId   string `json:"userId" dynamodbav:"userId"` // primary key
	Username string `json:"username" dynamodbav:"username"`
	Email    string `json:"email" dynamodbav:"email"` // global secondary index
	Active   bool   `json:"active" dynamodbav:"active"`
}

type PaymentResponse struct {
	PaymentId string    `json:"paymentId"`
	ExpenseId string    `json:"expenseId"`
	Date      time.Time `json:"date"`
	Status    string    `json:"status"`
}
