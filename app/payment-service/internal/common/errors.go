package common

import (
	"errors"
)

var (
	ErrOrderNotFound    = errors.New("ERROR : order not found")
	ErrOrderNoCreated   = errors.New("ERROR : operation failed")
	ErrUnespectedError  = errors.New("unespected error")
	ErrWrongCredentials = errors.New("ERROR : wrong credentials")
)
