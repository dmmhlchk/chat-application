package domain

import "time"

type Participant struct {
	ChatID   string
	UserID   string
	JoinedAt time.Time
}

func NewParticipant(
	chatID string,
	userID string,
) *Participant {
	now := time.Now().UTC()

	return &Participant{
		ChatID:   chatID,
		UserID:   userID,
		JoinedAt: now,
	}
}
