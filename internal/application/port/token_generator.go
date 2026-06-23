package port

import (
	"time"
)

type TokenGenerator interface {
	GenerateToken(userID string, sessionID string, ttl time.Duration) (string, error)
	ValidateToken(token string) (string, string, error) // return UserID and SessionID else error
}
