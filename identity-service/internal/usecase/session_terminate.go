package usecase

import (
	"context"
	"errors"
	"fmt"
	"identity-service/internal/domain"
)

// 1. Determine Input
// 1.1. TerminateSessionInput defines the data required to terminate the session
type TerminateSessionInput struct {
	UserID       int
	RefreshToken string
}

// 2. Determine the dependencies
// 2.1. TerminateSession coordinates domain layer to sign in a user
type TerminateSession struct {
	sessionRepo domain.SessionRepo
}

// 2.2. NewTerminateSession is a constructor that handles dependency injection
func NewTerminateSession(sessionRepo domain.SessionRepo) *TerminateSession {
	return &TerminateSession{
		sessionRepo: sessionRepo,
	}
}

// 3. Execute runs the actual step-by-step session termination business flow
func (ts *TerminateSession) Execute(ctx context.Context, input TerminateSessionInput) error {
	// 1. Validate input basic constraints
	if input.UserID == 0 || input.RefreshToken == "" {
		return errors.New("required fields were not filled")
	}

	session, err := ts.sessionRepo.FindByToken(ctx, input.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to lookup session: %w", err)
	}
	if session == nil {
		return errors.New("session not found or already terminated")
	}
	if session.UserID != input.UserID {
		return errors.New("unauthorized: you do not own this session")
	}

	err = ts.sessionRepo.DeleteByToken(ctx, input.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}
