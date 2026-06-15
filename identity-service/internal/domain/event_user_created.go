package domain

import "github.com/google/uuid"

type UserCreated struct {
	ID       uuid.UUID
	Username string
	Phone    string
}
