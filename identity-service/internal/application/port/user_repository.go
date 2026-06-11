package port

import (
	"context"

	"identity-service/internal/domain"

	"github.com/google/uuid"
)

type UserReader interface {
	FindByUserID(ctx context.Context, userID uuid.UUID) (*domain.User, error)
	FindByPhone(ctx context.Context, phone string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	ExistsByPhoneOrUsername(ctx context.Context, phone string, username string) (bool, error)
}

type UserWriter interface {
	Create(ctx context.Context, user *domain.User) error
	Update(ctx context.Context, user *domain.User) error
	Delete(ctx context.Context, userID uuid.UUID) error
}

type UserRepository interface {
	UserReader
	UserWriter
}
