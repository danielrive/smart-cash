package common

import (
	"errors"
)

var (
	ErrTransactionNotFound  = errors.New("ERROR : transaction not found")
	ErrUserNotFound         = errors.New("ERROR : user not found")
	ErrTransactionNoCreated = errors.New("ERROR : operation failed")
	ErrUnespectedError      = errors.New("unespected error")
	ErrWrongCredentials     = errors.New("ERROR : wrong credentials")
	ErrUserNoUpdated        = errors.New("ERROR : user not updated")
)
