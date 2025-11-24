package dto

type ProcessPaymentRequest struct {
	UserId    string  `json:"userId" validate:"required,uuid"`
	ExpenseId string  `json:"expenseId" validate:"required,uuid"`
	Amount    float64 `json:"amount" validate:"required,gt=0"`
}

