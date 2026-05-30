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
	ID                int
	UserID            int
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
	userId int,
	refreshTokenHash string,
	notificationToken string,
	device Device,
	ipAddress string,
	ttl time.Duration,
) *Session {
	now := time.Now().UTC()

	return &Session{
		UserID:            userId,
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

// IsValid checks if the session is active and not expired
func (s *Session) IsValid() bool {
	return !s.IsRevoked && !time.Now().UTC().After(s.ExpiresAt)
}

// Revoke changes the session state to revoked
func (s *Session) Revoke() {
	s.IsRevoked = true
}

type SessionRepo interface {
	FindByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*Session, error)
	FindAllByUserId(ctx context.Context, userId int) ([]*Session, error)

	Create(ctx context.Context, session *Session) error
	Update(ctx context.Context, session *Session) error

	DeleteAllByUserID(ctx context.Context, userId int) error
}
