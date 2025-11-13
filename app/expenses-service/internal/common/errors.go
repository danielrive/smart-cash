package common

import (
	"errors"
)

var (
	ErrExpenseNotFound        = errors.New("expenses not found")
	ErrExpenseNotPaid         = errors.New("expenses not paid")
	ErrExpenseNotCreated       = errors.New("operation failed")
	ErrUnexpectedError        = errors.New("unexpected error")
	ErrInternalError          = errors.New("internal error")
	ErrWrongCredentials       = errors.New("wrong credentials")
	ErrInsufficientFundsError = errors.New("insufficient funds")
	ErrUserNotFound           = errors.New("user not found")
)

type ErrorCode string

const (
	ErrNotFound     ErrorCode = "NOT_FOUND"
	ErrInvalidInput ErrorCode = "INVALID_INPUT"
	ErrInternal     ErrorCode = "INTERNAL_ERROR"
	ErrUnauthorized ErrorCode = "UNAUTHORIZED"
)

type AppError struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
	Details any       `json:"details,omitempty"`
}
