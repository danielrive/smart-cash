package common

import (
	"errors"
)

var (
	ErrExpenseNotFound        = errors.New("expenses not found")
	ErrExpenseNotPaid         = errors.New("expenses not paid")
	ErrExpenseNoCreated       = errors.New("operation failed")
	ErrUnespectedError        = errors.New("unespected error")
	ErrInternalError          = errors.New("internal error")
	ErrWrongCredentials       = errors.New("wrong credentials")
	ErrInsufficientFundsError = errors.New("insufficient funds")
	ErrUserNotFound           = errors.New("user not found")
)
