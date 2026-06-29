package security

import "time"

type TokenGenerator interface {
	Generate(userID string, sessionID string, ttl time.Duration) (string, error)
	Validate(token string) (string, string, error) // return UserID and SessionID else error
}
