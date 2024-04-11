package utils

import uuid "github.com/satori/go.uuid"

type UUIDHelper struct{}

func NewUUIDHelper() *UUIDHelper {
	return &UUIDHelper{}
}

func (u *UUIDHelper) New() string {
	return uuid.NewV4().String()
}
