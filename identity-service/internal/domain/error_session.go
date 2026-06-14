package domain

import "errors"

var (
	ErrSessionNotFound       = errors.New("session not found")
	ErrSessionAlreadyRevoked = errors.New("session already revoked")
	ErrSessionInvalid        = errors.New("session is invalid")
	ErrInvalidUserID         = errors.New("invalid user id")
	ErrAlreadyCleanSessions  = errors.New("user doesn't have any active session")
	ErrInvalidRefreshToken   = errors.New("invalid refresh token hash")
	ErrInvalidTTL            = errors.New("invalid ttl")
)
