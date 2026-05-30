package domain

import (
	"context"
	"time"
)

type User struct {
	ID           int
	Username     string
	Phone        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func NewUser(id int, username, phone, passwordHash string) *User {
	return &User{
		ID:           id,
		Username:     username,
		Phone:        phone,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

type UserRepo interface {
	FindByID(ctx context.Context, id int) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)

	ExistsByPhoneOrUsername(ctx context.Context, phone, username string) (bool, error)

	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, user *User) error
}
