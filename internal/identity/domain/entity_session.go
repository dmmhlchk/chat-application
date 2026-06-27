package domain

import "time"

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
