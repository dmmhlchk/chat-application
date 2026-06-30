package domain

import "time"

type Message struct {
	ID      string
	ChatID  string
	UserID  string
	Type    MessageType
	SentAt  time.Time
	Content string
}

func NewMessage(
	messageID string,
	chatID string,
	userID string,
	messageType MessageType,
	content string,
) *Message {
	now := time.Now().UTC()

	return &Message{
		ID:      messageID,
		ChatID:  chatID,
		UserID:  userID,
		Type:    messageType,
		SentAt:  now,
		Content: content,
	}
}
