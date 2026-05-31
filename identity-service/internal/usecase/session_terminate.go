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
	tokenGen    domain.TokenGenerator
}

func NewTerminateSession(sessionRepo domain.SessionRepo, tokenGen domain.TokenGenerator) *TerminateSession {
	return &TerminateSession{
		sessionRepo: sessionRepo,
		tokenGen:    tokenGen,
	}
}

// 3. Business flow of session termination
func (uc *TerminateSession) Execute(ctx context.Context, input TerminateSessionInput) error {
	// 1. Validate input basic constraints
	if input.UserID == 0 || input.RefreshToken == "" {
		return errors.New("required fields were not filled")
	}

	// 2. Retrive a token ID
	sessionId, userId, err := uc.tokenGen.ValidateRefreshTokenHash(input.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to validate a token: %w", err)
	}
	if userId != input.UserID {
		return errors.New("unauthorized: you do not own this session")
	}

	// 3. Find a token in DB and check if we can terminate the session or not
	session, err := uc.sessionRepo.FindByID(ctx, sessionId)
	if err != nil {
		return fmt.Errorf("failed to lookup session: %w", err)
	}
	if session == nil {
		return errors.New("session not found or already terminated")
	}
	if session.UserID != input.UserID {
		return errors.New("unauthorized: you do not own this session")
	}

	// 4. Revoke the session
	session.Revoke()
	err = uc.sessionRepo.Update(ctx, session)
	if err != nil {
		return fmt.Errorf("failed to revoke session: %w", err)
	}

	return nil
}
