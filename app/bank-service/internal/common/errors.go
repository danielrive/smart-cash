package common

import (
	"errors"
)

var (
	ErrUserNotFound           = errors.New("user not found")
	ErrTransactionFailed      = errors.New("transaction failed")
	ErrUnespectedError        = errors.New("unespected error")
	ErrInternalError          = errors.New("internal error")
	ErrWrongCredentials       = errors.New("wrong credentials")
	ErrInsufficientFundsError = errors.New("insufficient funds")
)
