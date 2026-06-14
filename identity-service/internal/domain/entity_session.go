package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID                uuid.UUID
	UserID            uuid.UUID
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
	sessionID uuid.UUID,
	userID uuid.UUID,
	refreshTokenHash string,
	notificationToken string,
	device Device,
	ipAddress string,
	ttl time.Duration,
) (*Session, error) {
	if userID == uuid.Nil {
		return nil, ErrInvalidUserID
	}
	if refreshTokenHash == "" {
		return nil, ErrInvalidRefreshToken
	}
	if ttl <= 0 {
		return nil, ErrInvalidTTL
	}

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
	}, nil
}

func (s *Session) IsExpired() bool {
	return time.Now().UTC().After(s.ExpiresAt)
}

func (s *Session) IsValid() bool {
	return !s.IsRevoked && !s.IsExpired()
}

func (s *Session) Revoke() error {
	if s.IsRevoked {
		return ErrSessionAlreadyRevoked
	}
	s.IsRevoked = true
	return nil
}

func (s *Session) RefreshActivity(ipAddress string) error {
	if !s.IsValid() {
		return ErrSessionInvalid
	}
	s.ActiveAt = time.Now().UTC()
	s.ActiveIPAddress = ipAddress
	return nil
}
