package domain

import (
	"context"
	"time"
)

type Device struct {
	Hash     string
	Name     string
	Version  string
	Platform Platform
}

type Session struct {
	ID                string
	UserID            string
	RefreshTokenHash  string
	NotificationToken string
	Device            Device
	CreatedAt         time.Time
	CreatedIPAddress  string
	ActiveAt          time.Time
	ActiveIPAddress   string
	ExpiresAt         time.Time
	IsRevoked         bool
}

func NewSession(
	sessionID string,
	userID string,
	refreshTokenHash string,
	notificationToken string,
	device Device,
	ipAddress string,
	ttl time.Duration,
) *Session {
	now := time.Now().UTC()

	return &Session{
		ID:                sessionID,
		UserID:            userID,
		RefreshTokenHash:  refreshTokenHash,
		NotificationToken: notificationToken,
		Device:            device,
		CreatedAt:         now,
		CreatedIPAddress:  ipAddress,
		ActiveAt:          now,
		ActiveIPAddress:   ipAddress,
		ExpiresAt:         now.Add(ttl),
		IsRevoked:         false,
	}
}

type SessionRepository interface {
	FindByID(ctx context.Context, sessionID string) (*Session, error) // return active session
	FindAll(ctx context.Context, userID string) ([]Session, error)    // return active sessions

	NewUUID() string

	TerminateByID(ctx context.Context, sessionID string) error
	TerminateAll(ctx context.Context, userID string) error

	Create(ctx context.Context, session *Session) error
	Update(ctx context.Context, session *Session) error
}
