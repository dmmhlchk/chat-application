package domain

import (
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
) *User {
	return &User{
		ID:           userID,
		Username:     username,
		Phone:        phone,
		PasswordHash: passwordHash,
	}
}
