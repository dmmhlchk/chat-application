package domain

import "time"

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
