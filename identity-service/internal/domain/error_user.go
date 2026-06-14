package domain

import "errors"

var (
	// user errors
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPhone      = errors.New("invalid phone number")
)
