package port

import (
	"context"

	"identity-service/internal/domain"

	"github.com/google/uuid"
)

type SessionReader interface {
	FindAllByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error)
	FindBySessionID(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error)
}

type SessionWriter interface {
	Create(ctx context.Context, session *domain.Session) error
	Update(ctx context.Context, session *domain.Session) error
	TerminateAllByUserID(ctx context.Context, userID uuid.UUID) error
	TerminateBySessionID(ctx context.Context, sessionID uuid.UUID) error
}

type SessionRepository interface {
	SessionReader
	SessionWriter
}
