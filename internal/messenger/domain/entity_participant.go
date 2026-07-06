package domain

import "time"

type Participant struct {
	ChatID   string
	UserID   string
	JoinedAt time.Time
	Tag      string
}

func NewParticipant(
	chatID string,
	userID string,
	tag string,
) *Participant {
	now := time.Now().UTC()

	return &Participant{
		ChatID:   chatID,
		UserID:   userID,
		JoinedAt: now,
		Tag:      tag,
	}
}
