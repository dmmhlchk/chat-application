package usecase

import (
	"context"
	"errors"
	"fmt"

	"identity-service/internal/application/port"

	"github.com/google/uuid"
)

// 1. Determine the input
type TerminateSessionInput struct {
	UserID       uuid.UUID
	RefreshToken string
}

// 2. Determine the dependencies
type TerminateSession struct {
	sessionRepo port.SessionRepository
	tokenGen    port.TokenGenerator
}

func NewTerminateSession(
	sessionRepo port.SessionRepository,
	tokenGen port.TokenGenerator,
) *TerminateSession {
	return &TerminateSession{
		sessionRepo: sessionRepo,
		tokenGen:    tokenGen,
	}
}

// 3. Business flow of session termination
func (uc *TerminateSession) Execute(ctx context.Context, input TerminateSessionInput) error {
	// 1. Validate input basic constraints
	if input.UserID == uuid.Nil || input.RefreshToken == "" {
		return errors.New("required fields were not filled")
	}

	// 2. Retrive a token ID
	userID, sessionID, err := uc.tokenGen.ValidateToken(input.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to validate a token: %w", err)
	}
	if userID != input.UserID {
		return errors.New("unauthorized: you do not own this session")
	}

	// 3. Find a token in DB and check if we can terminate the session or not
	session, err := uc.sessionRepo.FindBySessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to lookup session: %w", err)
	}
	if session == nil {
		return errors.New("session not found or already terminated")
	}
	if session.ID != sessionID {
		return errors.New("unauthorized: you do not own this session")
	}

	// 4. Revoke the session
	err = uc.sessionRepo.TerminateBySessionID(ctx, sessionID)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}
