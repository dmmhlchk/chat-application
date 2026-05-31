package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidUsername = errors.New("username cannot be empty")
	ErrInvalidPhone    = errors.New("invalid phone number format")
)

type User struct {
	ID           int
	Username     string
	Phone        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(id int, username, phone, passwordHash string) (*User, error) {
	return &User{
		ID:           id,
		Username:     username,
		Phone:        phone,
		PasswordHash: passwordHash,
	}, nil
}

type UserRepo interface {
	ExistsByPhoneOrUsername(ctx context.Context, phone, username string) (bool, error)

	FindByID(ctx context.Context, id int) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)

	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id int) error
}
