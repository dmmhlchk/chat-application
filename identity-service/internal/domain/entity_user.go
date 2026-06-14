package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID           uuid.UUID
	Username     string
	Phone        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(
	userID uuid.UUID,
	username string,
	phone string,
	passwordHash string,
) (*User, error) {
	return &User{
		ID:           userID,
		Username:     username,
		Phone:        phone,
		PasswordHash: passwordHash,
	}, nil
}
