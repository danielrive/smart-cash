package common

import (
	"errors"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserNotCreated    = errors.New("user not created")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInternalError     = errors.New("internal error")
	ErrUnexpectedError   = errors.New("unexpected error")
	ErrWrongCredentials  = errors.New("wrong credentials")
)
