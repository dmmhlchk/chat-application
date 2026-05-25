package domain

import "time"

type Device struct {
	Hash     string
	Name     string
	Version  string
	Platform Platform
}

type Session struct {
	ID                int
	UserID            int
	RefreshToken      string
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
	refreshToken string,
	notificationToken string,
	device Device,
	ipAddress string,
	ttl time.Duration,
) *Session {
	now := time.Now()

	return &Session{
		UserID:            userId,
		RefreshToken:      refreshToken,
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

func (s *Session) IsValid() bool {
	return !s.IsRevoked && time.Now().Before(s.ExpiresAt)
}

func (s *Session) Revoke() {
	s.IsRevoked = true
}
