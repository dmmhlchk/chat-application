package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"identity-service/internal/application/usecase"
	"identity-service/internal/domain"
)

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestTerminateSession_Success(t *testing.T) {
	ctx := context.Background()

	sessionRepo := &mockSessionRepository{}
	tokenGen := &mockTokenGenerator{}

	session := &domain.Session{ID: "sess-001"}

	tokenGen.On("ValidateToken", "valid-refresh-token").
		Return("user-001", "sess-001", nil)
	sessionRepo.On("FindBySessionID", ctx, "sess-001").Return(session, nil)
	sessionRepo.On("TerminateBySessionID", ctx, "sess-001").Return(nil)

	uc := usecase.NewTerminateSession(sessionRepo, tokenGen)
	err := uc.Execute(ctx, usecase.TerminateSessionInput{
		UserID:       "user-001",
		RefreshToken: "valid-refresh-token",
	})

	assert.NoError(t, err)
	sessionRepo.AssertExpectations(t)
	tokenGen.AssertExpectations(t)
}

func TestTerminateSession_MissingRequiredFields(t *testing.T) {
	cases := []struct {
		name  string
		input usecase.TerminateSessionInput
	}{
		{
			name:  "empty userID",
			input: usecase.TerminateSessionInput{RefreshToken: "some-token"},
		},
		{
			name:  "empty refreshToken",
			input: usecase.TerminateSessionInput{UserID: "user-001"},
		},
		{
			name:  "both empty",
			input: usecase.TerminateSessionInput{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sessionRepo := &mockSessionRepository{}
			tokenGen := &mockTokenGenerator{}

			uc := usecase.NewTerminateSession(sessionRepo, tokenGen)
			err := uc.Execute(context.Background(), tc.input)

			assert.EqualError(t, err, "required fields were not filled")
			tokenGen.AssertNotCalled(t, "ValidateToken", mock.Anything)
		})
	}
}

func TestTerminateSession_InvalidToken(t *testing.T) {
	ctx := context.Background()

	sessionRepo := &mockSessionRepository{}
	tokenGen := &mockTokenGenerator{}

	tokenGen.On("ValidateToken", "bad-token").
		Return("", "", errors.New("token expired"))

	uc := usecase.NewTerminateSession(sessionRepo, tokenGen)
	err := uc.Execute(ctx, usecase.TerminateSessionInput{
		UserID:       "user-001",
		RefreshToken: "bad-token",
	})

	assert.ErrorContains(t, err, "failed to validate a token")
	sessionRepo.AssertNotCalled(t, "FindBySessionID", mock.Anything, mock.Anything)
}

func TestTerminateSession_TokenBelongsToDifferentUser(t *testing.T) {
	ctx := context.Background()

	sessionRepo := &mockSessionRepository{}
	tokenGen := &mockTokenGenerator{}

	// Token's embedded userID does not match input.UserID
	tokenGen.On("ValidateToken", "other-users-token").
		Return("user-999", "sess-001", nil)

	uc := usecase.NewTerminateSession(sessionRepo, tokenGen)
	err := uc.Execute(ctx, usecase.TerminateSessionInput{
		UserID:       "user-001",
		RefreshToken: "other-users-token",
	})

	assert.EqualError(t, err, "unauthorized: you do not own this session")
	sessionRepo.AssertNotCalled(t, "FindBySessionID", mock.Anything, mock.Anything)
}

func TestTerminateSession_SessionNotFound(t *testing.T) {
	ctx := context.Background()

	sessionRepo := &mockSessionRepository{}
	tokenGen := &mockTokenGenerator{}

	tokenGen.On("ValidateToken", "valid-refresh-token").
		Return("user-001", "sess-001", nil)
	sessionRepo.On("FindBySessionID", ctx, "sess-001").Return(nil, nil)

	uc := usecase.NewTerminateSession(sessionRepo, tokenGen)
	err := uc.Execute(ctx, usecase.TerminateSessionInput{
		UserID:       "user-001",
		RefreshToken: "valid-refresh-token",
	})

	assert.EqualError(t, err, "session not found or already terminated")
	sessionRepo.AssertNotCalled(t, "TerminateBySessionID", mock.Anything, mock.Anything)
}

func TestTerminateSession_FindBySessionIDError(t *testing.T) {
	ctx := context.Background()

	sessionRepo := &mockSessionRepository{}
	tokenGen := &mockTokenGenerator{}

	tokenGen.On("ValidateToken", "valid-refresh-token").
		Return("user-001", "sess-001", nil)
	sessionRepo.On("FindBySessionID", ctx, "sess-001").
		Return(nil, errors.New("db error"))

	uc := usecase.NewTerminateSession(sessionRepo, tokenGen)
	err := uc.Execute(ctx, usecase.TerminateSessionInput{
		UserID:       "user-001",
		RefreshToken: "valid-refresh-token",
	})

	assert.ErrorContains(t, err, "failed to lookup session")
	sessionRepo.AssertNotCalled(t, "TerminateBySessionID", mock.Anything, mock.Anything)
}

func TestTerminateSession_TerminateError(t *testing.T) {
	ctx := context.Background()

	sessionRepo := &mockSessionRepository{}
	tokenGen := &mockTokenGenerator{}

	session := &domain.Session{ID: "sess-001"}

	tokenGen.On("ValidateToken", "valid-refresh-token").
		Return("user-001", "sess-001", nil)
	sessionRepo.On("FindBySessionID", ctx, "sess-001").Return(session, nil)
	sessionRepo.On("TerminateBySessionID", ctx, "sess-001").
		Return(errors.New("db write error"))

	uc := usecase.NewTerminateSession(sessionRepo, tokenGen)
	err := uc.Execute(ctx, usecase.TerminateSessionInput{
		UserID:       "user-001",
		RefreshToken: "valid-refresh-token",
	})

	assert.ErrorContains(t, err, "failed to revoke session")
}
