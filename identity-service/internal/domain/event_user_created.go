package domain

import "github.com/google/uuid"

type UserCreated struct {
	UserID uuid.UUID
}
