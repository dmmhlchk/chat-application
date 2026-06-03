package domain

import (
	"context"
	"time"
)

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) (bool, error)
}

type TokenGenerator interface {
	GenerateToken(userID, sessionID string, ttl time.Duration) (string, error)
	ValidateToken(token string) (string, string, error) // return UserID and SessionID else error
}

type OTPGenerator interface {
	Generate(length int) (string, error)
}

type OTPRepository interface {
	Save(ctx context.Context, phone string, code string) error
	Verify(ctx context.Context, phone string, code string) (bool, error)
	Delete(ctx context.Context, phone string) error
}
