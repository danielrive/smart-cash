package common

import (
	"errors"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrUserNoCreated    = errors.New("user not created")
	ErrInternalError    = errors.New("internal error")
	ErrUnespectedError  = errors.New("unespected error")
	ErrWrongCredentials = errors.New("wrong credentials")
)
