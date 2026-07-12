package repository

import (
	"chat-app/internal/messenger/domain"
	"context"
)

type MessageReader interface {
	GetHistory(ctx context.Context, chatID string, page int) ([]domain.Message, error)
}

type MessageWriter interface {
	Send(ctx context.Context, message *domain.Message) error
}

type MessageRepositroy interface {
	MessageReader
	MessageWriter
}
