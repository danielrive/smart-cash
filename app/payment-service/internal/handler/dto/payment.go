package dto

type ProcessPaymentRequest struct {
	ExpenseId string  `json:"expenseId" validate:"required,uuid"`
	Amount    float64 `json:"amount" validate:"required,gt=0"`
}
