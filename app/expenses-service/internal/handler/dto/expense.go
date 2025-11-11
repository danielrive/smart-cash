// internal/handler/dto/expense.go
package dto

type CreateExpenseRequest struct {
	UserId      string   `json:"-"` // Not from JSON, set from auth middleware
	Name        string   `json:"name" validate:"required,max=100"`
	Amount      float64  `json:"amount" validate:"required,gt=0"`
	Description string   `json:"description" validate:"required,max=500"`
	Category    string   `json:"category" validate:"required"`
	Date        string   `json:"date" validate:"required,datetime=2006-01-02"`
	Tags        []string `json:"tags,omitempty" validate:"omitempty,dive,max=50"`
}

type UpdateExpenseRequest struct {
	Name        *string  `json:"name,omitempty" validate:"omitempty,max=100"`
	Amount      *float64 `json:"amount,omitempty" validate:"omitempty,gt=0"`
	Description *string  `json:"description,omitempty" validate:"omitempty,max=500"`
	Category    *string  `json:"category,omitempty"`
	Tags        []string `json:"tags,omitempty" validate:"omitempty,dive,max=50"`
}
