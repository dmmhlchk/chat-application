package port

import (
	"time"

	"github.com/google/uuid"
)

type TokenGenerator interface {
	GenerateToken(userID uuid.UUID, sessionID uuid.UUID, ttl time.Duration) (string, error)
	ValidateToken(token string) (uuid.UUID, uuid.UUID, error) // return UserID and SessionID else error
}
