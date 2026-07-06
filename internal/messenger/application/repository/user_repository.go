package repository

import (
	"context"

	"chat-app/internal/identity/domain"
)

type UserReader interface {
	FindByUserID(ctx context.Context, userID string) (*domain.User, error)
	ExistsByUserID(ctx context.Context, userID string) (bool, error)
	ExistsByUserIDs(ctx context.Context, userIDs ...string) (bool, error)
}

type UserWriter interface {
}

type UserRepository interface {
	UserReader
	UserWriter
}
