package usecase

import (
	"context"
	"errors"
	"fmt"
	"identity-service/internal/domain"
)

// 1. Determine the input
type TerminateSessionInput struct {
	UserID       int
	RefreshToken string
}

// 2. Determine the dependencies
type TerminateSession struct {
	sessionRepo domain.SessionRepo
}

func NewTerminateSession(sessionRepo domain.SessionRepo) *TerminateSession {
	return &TerminateSession{
		sessionRepo: sessionRepo,
	}
}

// 3. Business flow of session termination
func (uc *TerminateSession) Execute(ctx context.Context, input TerminateSessionInput) error {
	// 1. Validate input basic constraints
	if input.UserID == 0 || input.RefreshToken == "" {
		return errors.New("required fields were not filled")
	}

	session, err := uc.sessionRepo.FindByRefreshTokenHash(ctx, input.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to lookup session: %w", err)
	}
	if session == nil {
		return errors.New("session not found or already terminated")
	}
	if session.UserID != input.UserID {
		return errors.New("unauthorized: you do not own this session")
	}

	session.Revoke()
	err = uc.sessionRepo.Update(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}
