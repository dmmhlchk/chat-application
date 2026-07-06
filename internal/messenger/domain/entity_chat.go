package domain

import "time"

type Chat struct {
	ID        string
	Type      ChatType
	CreatedAt time.Time
	Title     string
}

type ChatOption func(*Chat)

func WithTitle(title string) ChatOption {
	return func(c *Chat) {
		c.Title = title
	}
}

func NewChat(
	chatID string,
	chatType ChatType,
	chatOpts ...ChatOption,
) *Chat {
	now := time.Now().UTC()

	c := &Chat{
		ID:        chatID,
		Type:      chatType,
		CreatedAt: now,
	}

	for _, opt := range chatOpts {
		opt(c)
	}

	return c
}
