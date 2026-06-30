package domain

import "time"

type Chat struct {
	ID        string
	Type      ChatType
	CreatedAt time.Time
	Title     string
}

func NewChat(
	chatID string,
	chatType ChatType,
	title string,
) *Chat {
	now := time.Now().UTC()

	return &Chat{
		ID:        chatID,
		Type:      chatType,
		CreatedAt: now,
		Title:     title,
	}
}
