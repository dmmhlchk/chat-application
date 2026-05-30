package domain

import "time"

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) (bool, error)
}

type TokenGenerator interface {
	GeneratePair(UserID int, ttl time.Duration) (string, string, error)

	GenerateRefreshToken(userID int, ttl time.Duration) (string, error)
	ValidateRefreshToken(token string) (int, error) // return UserID else error

	GenerateAccessToken(userID int, ttl time.Duration) (string, error)
	ValidateAccessToken(token string) (int, error) // return UserID else error
}
