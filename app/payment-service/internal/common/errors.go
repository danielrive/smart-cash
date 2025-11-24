package common

import (
	"errors"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrExpenseNotFound        = errors.New("expense not found")
	ErrPaymentNotFound        = errors.New("payment not found")
	ErrPaymentFailed          = errors.New("payment failed")
	ErrUnexpectedError        = errors.New("unexpected error")
	ErrInternalError          = errors.New("internal error")
	ErrWrongCredentials       = errors.New("wrong credentials")
	ErrInsufficientFundsError = errors.New("insufficient funds")
	ErrUserBlocked            = errors.New("user is blocked, contact support center")
)
