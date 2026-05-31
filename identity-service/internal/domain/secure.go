package domain

import (
	"context"
	"time"
)

// PasswordHasher protects user's passwords and validates them
type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) (bool, error)
}

// TokenGenerator handles identity tokens
type TokenGenerator interface {
	GeneratePair(UserID int, ttl time.Duration) (string, string, error)

	GenerateRefreshTokenHash(userID int, ttl time.Duration) (string, error)
	ValidateRefreshTokenHash(token string) (int, int, error) // return SessionID and UserID else error

	GenerateAccessToken(userID int, ttl time.Duration) (string, error)
	ValidateAccessToken(token string) (int, int, error) // return SessionID and UserID else error
}

// OTPGenerator generates an un-guessable string of digits
type OTPGenerator interface {
	Generate(length int) (string, error)
}

// CodeHandler handles short-lived OTP transient data
type CodeHandler interface {
	Save(ctx context.Context, phone string, code string) error
	Verify(ctx context.Context, phone string, code string) (bool, error)
	Delete(ctx context.Context, phone string) error
}
