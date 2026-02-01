package dto

type ProcessPaymentRequest struct {
	ExpenseId string  `json:"expenseId" validate:"required,uuid"`
	Amount    float64 `json:"amount" validate:"required,gt=0"`
}

type UpdatePaymentStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending processing completed failed cancelled"`
}
