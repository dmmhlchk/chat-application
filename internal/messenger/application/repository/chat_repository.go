package repository

import (
	"context"

	"chat-app/internal/messenger/domain"
)

type ChatMembership interface {
	IsMember(ctx context.Context, userID string, chatID string) (bool, error)
}

type ChatPermissions interface {
	CheckPermissions(ctx context.Context, chatID string, userID string, permission string) (bool, error)
	GetUserPermissions(ctx context.Context, chatID string, userID string) (string, error)
}

type ChatReader interface {
	FindByChatID(ctx context.Context, chatID string) (*domain.Chat, error)
	FindDirectBetween(ctx context.Context, userIDs ...string) (*domain.Chat, error)
}

type ChatWriter interface {
	Create(ctx context.Context, chat *domain.Chat) error
	Update(ctx context.Context, chat *domain.Chat) error
	Delete(ctx context.Context, chatID string) error

	Join(ctx context.Context, chatID string, userIDs ...string) error
	Leave(ctx context.Context, chatID string, userIDs ...string) error
	SetOwner(ctx context.Context, chatID string, userID string) error
}

type ChatRepository interface {
	ChatMembership
	ChatPermissions
	ChatReader
	ChatWriter
}
