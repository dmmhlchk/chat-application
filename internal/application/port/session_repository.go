package port

import (
	"chat-application/internal/domain"
	"context"
)

type SessionReader interface {
	FindAllByUserID(ctx context.Context, userID string) ([]domain.Session, error)
	FindBySessionID(ctx context.Context, sessionID string) (*domain.Session, error)
}

type SessionWriter interface {
	Create(ctx context.Context, session *domain.Session) error
	Update(ctx context.Context, session *domain.Session) error
	TerminateAllByUserID(ctx context.Context, userID string) error
	TerminateBySessionID(ctx context.Context, sessionID string) error
}

type SessionRepository interface {
	SessionReader
	SessionWriter
}
