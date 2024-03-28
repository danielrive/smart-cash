package common

import (
	"errors"
)

var (
	ErrUserNotFound     = errors.New("ERROR : user not found")
	ErrUserNoCreated    = errors.New("ERROR : operation failed")
	ErrUnespectedError  = errors.New("unespected error")
	ErrWrongCredentials = errors.New("ERROR : wrong credentials")
)
