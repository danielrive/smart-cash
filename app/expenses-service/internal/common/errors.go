package common

import (
	"errors"
)

var (
	ErrExpenseNotFound  = errors.New("ERROR : expenses not found")
	ErrExpenseNoCreated = errors.New("ERROR : operation failed")
	ErrUnespectedError  = errors.New("unespected error")
	ErrWrongCredentials = errors.New("ERROR : wrong credentials")
)
