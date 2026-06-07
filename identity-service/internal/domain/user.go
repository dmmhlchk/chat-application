package domain

import (
	"context"
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

func NewUser(userID, username, phone, passwordHash string) (*User, error) {
	return &User{
		ID:           userID,
		Username:     username,
		Phone:        phone,
		PasswordHash: passwordHash,
	}, nil
}

type UserRepository interface {
	ExistsByPhoneOrUsername(ctx context.Context, phone, username string) (bool, error)

	FindByID(ctx context.Context, userID string) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)

	NewUUID() string

	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, userID string) error
}
