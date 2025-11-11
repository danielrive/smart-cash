package dto

// PayExpenseRequest represents the request body for paying an expense
type PayExpenseRequest struct {
	TransactionId string  `json:"transactionId" validate:"required,uuid"`
	ExpenseId     string  `json:"expenseId" validate:"required,uuid"`
	Date          string  `json:"date" validate:"required,datetime=2006-01-02"`
	Amount        float64 `json:"amount" validate:"required,gt=0"`
	UserId        string  `json:"userId" validate:"required,uuid"`
	Status        string  `json:"status" validate:"required,oneof=pending completed failed"`
}

