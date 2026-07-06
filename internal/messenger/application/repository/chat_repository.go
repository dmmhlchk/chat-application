package repository

import (
	"context"

	"chat-app/internal/messenger/domain"
)

type ChatMembership interface {
	IsMember(ctx context.Context, userID string, chatID string) (bool, error)
}

type ChatExistence interface {
	ExistsByChatID(ctx context.Context, chatID string) (bool, error)
	ExistsDirectBetween(ctx context.Context, userIDs ...string) (bool, error)
}

type ChatReader interface {
	FindByChatID(ctx context.Context, chatID string) (*domain.Chat, error)
}

type ChatWriter interface {
	Create(ctx context.Context, chat *domain.Chat) error
	Update(ctx context.Context, chat *domain.Chat) error
	Delete(ctx context.Context, chatID string) error
}

type ChatRepository interface {
	ChatMembership
	ChatExistence
	ChatReader
	ChatWriter
}
