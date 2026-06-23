package domain

import (
	"strings"
	"time"
)

type User struct {
	ID           string
	Username     string
	Phone        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(
	userID string,
	username string,
	phone string,
	passwordHash string,
) (*User, error) {
	if strings.TrimSpace(username) == "" {
		return nil, ErrInvalidUsername
	}
	if strings.TrimSpace(phone) == "" {
		return nil, ErrInvalidPhone
	}

	return &User{
		ID:           userID,
		Username:     username,
		Phone:        phone,
		PasswordHash: passwordHash,
	}, nil
}
