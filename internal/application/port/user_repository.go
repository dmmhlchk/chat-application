package port

import (
	"context"
	"internal/domain"
)

type UserReader interface {
	FindByUserID(ctx context.Context, userID string) (*domain.User, error)
	FindByPhone(ctx context.Context, phone string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	ExistsByPhoneOrUsername(ctx context.Context, phone string, username string) (bool, error)
}

type UserWriter interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, userID string) error
}

type UserRepository interface {
	UserReader
	UserWriter
}
