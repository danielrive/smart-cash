package common

import (
	"errors"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrExpenseNotFound        = errors.New("expense not found")
	ErrTransactionNotFound    = errors.New("transaction not found")
	ErrTransactionFailed      = errors.New("transaction failed")
	ErrUnexpectedError        = errors.New("unexpected error")
	ErrInternalError          = errors.New("internal error")
	ErrWrongCredentials       = errors.New("wrong credentials")
	ErrInsufficientFundsError = errors.New("insufficient funds")
	ErrUserBlocked            = errors.New("user is blocked, contact support center")
)
