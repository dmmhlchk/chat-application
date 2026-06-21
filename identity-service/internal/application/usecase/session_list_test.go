package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"identity-service/internal/application/usecase"
	"identity-service/internal/domain"
)

// ___ Mock _________________________________________________________________

type mockSessionReader struct{ mock.Mock }

func (m *mockSessionReader) FindAllByUserID(ctx context.Context, userID string) ([]domain.Session, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Session), args.Error(1)
}

func (m *mockSessionReader) FindBySessionID(ctx context.Context, sessionID string) (*domain.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

// ___ Tests _________________________________________________________________

func TestSessionList_Success_MarksCurrentSession(t *testing.T) {
	ctx := context.Background()
	reader := &mockSessionReader{}

	now := time.Now()
	sessions := []domain.Session{
		{
			ID:               "sess-001",
			RefreshTokenHash: "token-hash-current",
			ActiveIPAddress:  "10.0.0.1",
			Device: domain.Device{
				Hash:     "c191537f-df38-417b-ad14-00b6d10117b1",
				Name:     "iPhone 15",
				Version:  "17",
				Platform: domain.PlatformIOS,
			},
			CreatedAt: now,
		},
		{
			ID:               "sess-002",
			RefreshTokenHash: "token-hash-other",
			ActiveIPAddress:  "10.0.0.2",
			Device: domain.Device{
				Hash:     "4452e89c-3749-4f61-8eb8-cab5527916aa",
				Name:     "macOS 13",
				Version:  "13",
				Platform: domain.PlatformMacOS,
			},
			CreatedAt: now.Add(-24 * time.Hour),
		},
	}

	reader.On("FindAllByUserID", ctx, "user-001").Return(sessions, nil)

	uc := usecase.NewSessionList(reader)
	out, err := uc.Execute(ctx, usecase.SessionListInput{
		UserID:              "user-001",
		CurrentRefreshToken: "token-hash-current",
	})

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Len(t, out.Sessions, 2)

	// First session is the current one
	assert.Equal(t, "sess-001", out.Sessions[0].ID)
	assert.True(t, out.Sessions[0].IsCurrent)
	assert.Equal(t, "10.0.0.1", out.Sessions[0].IPAddress)
	assert.Equal(t,
		domain.Device{
			Hash:     "c191537f-df38-417b-ad14-00b6d10117b1",
			Name:     "iPhone 15",
			Version:  "17",
			Platform: domain.PlatformIOS,
		}, out.Sessions[0].Device)

	// Second session is not current
	assert.Equal(t, "sess-002", out.Sessions[1].ID)
	assert.False(t, out.Sessions[1].IsCurrent)
	assert.Equal(t, "10.0.0.2", out.Sessions[1].IPAddress)

	reader.AssertExpectations(t)
}

func TestSessionList_Success_EmptySessions(t *testing.T) {
	ctx := context.Background()
	reader := &mockSessionReader{}

	reader.On("FindAllByUserID", ctx, "user-001").Return([]domain.Session{}, nil)

	uc := usecase.NewSessionList(reader)
	out, err := uc.Execute(ctx, usecase.SessionListInput{
		UserID:              "user-001",
		CurrentRefreshToken: "some-token",
	})

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Empty(t, out.Sessions)
}

func TestSessionList_Success_NoSessionMatchesCurrentToken(t *testing.T) {
	ctx := context.Background()
	reader := &mockSessionReader{}

	sessions := []domain.Session{
		{ID: "sess-001", RefreshTokenHash: "hash-a"},
		{ID: "sess-002", RefreshTokenHash: "hash-b"},
	}
	reader.On("FindAllByUserID", ctx, "user-001").Return(sessions, nil)

	uc := usecase.NewSessionList(reader)
	out, err := uc.Execute(ctx, usecase.SessionListInput{
		UserID:              "user-001",
		CurrentRefreshToken: "hash-unknown",
	})

	assert.NoError(t, err)
	assert.Len(t, out.Sessions, 2)
	for _, s := range out.Sessions {
		assert.False(t, s.IsCurrent, "no session should be marked as current")
	}
}

func TestSessionList_RepoError(t *testing.T) {
	ctx := context.Background()
	reader := &mockSessionReader{}

	reader.On("FindAllByUserID", ctx, "user-001").
		Return([]domain.Session{}, errors.New("db connection lost"))

	uc := usecase.NewSessionList(reader)
	out, err := uc.Execute(ctx, usecase.SessionListInput{
		UserID:              "user-001",
		CurrentRefreshToken: "some-token",
	})

	assert.Nil(t, out)
	assert.ErrorContains(t, err, "failed to retrieve active sessions")
}
