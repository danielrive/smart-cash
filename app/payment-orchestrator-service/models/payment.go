package models

// PaymentEvent represents the payment event received from SQS
// This matches the structure published by the lambda-processor
type PaymentEvent struct {
	PaymentId string  `json:"paymentId"`
	UserId    string  `json:"userId"`
	ExpenseId string  `json:"expenseId"`
	Amount    float64 `json:"amount"`
	Status    string  `json:"status"`
}

// UpdateExpenseStatusRequest represents the request to update expense status
type UpdateExpenseStatusRequest struct {
	Status string `json:"status"`
}

// ProcessTransactionRequest represents the request to process a bank transaction
type ProcessTransactionRequest struct {
	UserId        string  `json:"userId"`
	Amount        float64 `json:"amount"`
	TransactionId string  `json:"transactionId"`
	ExpenseId     string  `json:"expenseId"`
}

// UpdatePaymentStatusRequest represents the request to update payment status
type UpdatePaymentStatusRequest struct {
	Status string `json:"status"`
}
