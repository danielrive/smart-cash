package dto

// CreateUserRequest represents the request body for creating a new user
type CreateUserRequest struct {
	FirstName string `json:"firstName" validate:"required,min=2,max=50"`
	LastName  string `json:"lastName" validate:"required,min=2,max=50"`
	Username  string `json:"username" validate:"required,min=3,max=30,alphanum"`
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8,max=100"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
}
