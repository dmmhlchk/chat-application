package repository

import (
	"context"

	"chat-app/internal/messenger/domain"
)

type ChatReader interface {
	FindByChatID(ctx context.Context, chatID string) (*domain.Chat, error)
	ExistsDirectChat(ctx context.Context, userIDs ...string) (bool, error) // true if direct chat already exists between two users
}

type ChatWriter interface {
	Create(ctx context.Context, chat *domain.Chat) error
}

type ChatRepository interface {
	ChatReader
	ChatWriter
}
