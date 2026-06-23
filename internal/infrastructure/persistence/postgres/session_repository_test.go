package postgres_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"
	"time"

	"internal/domain"
	"internal/infrastructure/persistence/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ------------------------------------------------------------------------------------------------
// --	Helpers
// ------------------------------------------------------------------------------------------------

func newSessionRepoMock(t *testing.T) (*postgres.SessionRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	return postgres.NewSessionRepository(db).(*postgres.SessionRepository), mock
}

func sessionColumns() []string {
	return []string{
		"id", "user_id",
		"refresh_token", "notification_token",
		"device_hash", "device_name", "device_version", "device_platform",
		"created_ip_address", "active_ip_address",
		"created_at", "active_at",
		"expires_at", "is_revoked",
	}
}

func mockSessionRow(s *domain.Session) []driver.Value {
	return []driver.Value{
		s.ID, s.UserID,
		s.RefreshTokenHash, s.NotificationToken,
		s.Device.Hash, s.Device.Name, s.Device.Version, s.Device.Platform,
		s.CreatedIPAddress, s.ActiveIPAddress,
		s.CreatedAt, s.ActiveAt,
		s.ExpiresAt, s.IsRevoked,
	}
}

func newTestSession(userID string) *domain.Session {
	now := time.Now()
	return &domain.Session{
		ID:                uuid.New().String(),
		UserID:            userID,
		RefreshTokenHash:  "hashed_refresh_token",
		NotificationToken: "fcm_token_abc123",
		Device: domain.Device{
			Hash:     "device_hash_xyz",
			Name:     "iPhone 15",
			Version:  "17.0",
			Platform: "ios",
		},
		CreatedIPAddress: "192.168.1.1",
		ActiveIPAddress:  "192.168.1.2",
		CreatedAt:        now,
		ActiveAt:         now,
		ExpiresAt:        now.Add(30 * 24 * time.Hour),
		IsRevoked:        false,
	}
}

// ------------------------------------------------------------------------------------------------
// --	FindAllByUserID
// ------------------------------------------------------------------------------------------------

func TestSessionRepository_FindAllByUserID(t *testing.T) {
	t.Run("returns multiple sessions", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)
		userID := uuid.New().String()
		s1 := newTestSession(userID)
		s2 := newTestSession(userID)

		mock.ExpectQuery(`select .+ from sessions where user_id = \$1`).
			WithArgs(userID).
			WillReturnRows(
				sqlmock.NewRows(sessionColumns()).
					AddRow(mockSessionRow(s1)...).
					AddRow(mockSessionRow(s2)...),
			)

		sessions, err := repo.FindAllByUserID(context.Background(), userID)

		require.NoError(t, err)
		assert.Len(t, sessions, 2)
		assert.Equal(t, s1.ID, sessions[0].ID)
		assert.Equal(t, s2.ID, sessions[1].ID)
	})

	t.Run("returns empty slice when no sessions", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)

		mock.ExpectQuery(`select .+ from sessions where user_id = \$1`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows(sessionColumns())) // empty result

		sessions, err := repo.FindAllByUserID(context.Background(), uuid.New().String())

		require.NoError(t, err)
		// nil slice is acceptable here — callers should handle both nil and empty
		assert.Empty(t, sessions)
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)
		dbErr := errors.New("connection refused")

		mock.ExpectQuery(`select .+ from sessions where user_id = \$1`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(dbErr)

		sessions, err := repo.FindAllByUserID(context.Background(), uuid.New().String())

		assert.Nil(t, sessions)
		assert.ErrorIs(t, err, dbErr)
	})

	t.Run("scan error on malformed row", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)

		// returns a row with wrong column types to trigger scan failure
		mock.ExpectQuery(`select .+ from sessions where user_id = \$1`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnRows(
				sqlmock.NewRows(sessionColumns()).
					AddRow("not-a-uuid", nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil),
			)

		sessions, err := repo.FindAllByUserID(context.Background(), uuid.New().String())

		assert.Nil(t, sessions)
		assert.Error(t, err)
	})
}

// ------------------------------------------------------------------------------------------------
// --	FindBySessionID
// ------------------------------------------------------------------------------------------------

func TestSessionRepository_FindBySessionID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)
		s := newTestSession(uuid.New().String())

		mock.ExpectQuery(`select .+ from sessions where id = \$1`).
			WithArgs(s.ID).
			WillReturnRows(
				sqlmock.NewRows(sessionColumns()).AddRow(mockSessionRow(s)...),
			)

		got, err := repo.FindBySessionID(context.Background(), s.ID)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, s.ID, got.ID)
		assert.Equal(t, s.UserID, got.UserID)
		assert.Equal(t, s.Device.Name, got.Device.Name)
		assert.Equal(t, s.RefreshTokenHash, got.RefreshTokenHash)
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)

		mock.ExpectQuery(`select .+ from sessions where id = \$1`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(sql.ErrNoRows)

		got, err := repo.FindBySessionID(context.Background(), uuid.New().String())

		assert.Nil(t, got)
		assert.ErrorIs(t, err, domain.ErrSessionNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)
		dbErr := errors.New("timeout")

		mock.ExpectQuery(`select .+ from sessions where id = \$1`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(dbErr)

		got, err := repo.FindBySessionID(context.Background(), uuid.New().String())

		assert.Nil(t, got)
		assert.ErrorIs(t, err, dbErr)
	})
}

// ------------------------------------------------------------------------------------------------
// --	Create
// ------------------------------------------------------------------------------------------------

func TestSessionRepository_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)
		s := newTestSession(uuid.New().String())

		mock.ExpectExec(`insert into sessions`).
			WithArgs(
				s.UserID,
				s.RefreshTokenHash,
				s.NotificationToken,
				s.Device.Hash,
				s.Device.Name,
				s.Device.Version,
				s.Device.Platform,
				s.CreatedIPAddress,
				s.ActiveIPAddress,
				s.ExpiresAt,
			).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), s)

		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)
		dbErr := errors.New("unique constraint violation")

		mock.ExpectExec(`insert into sessions`).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(),
			).
			WillReturnError(dbErr)

		err := repo.Create(context.Background(), newTestSession(uuid.New().String()))

		assert.ErrorIs(t, err, dbErr)
	})
}

// ------------------------------------------------------------------------------------------------
// --	Update
// ------------------------------------------------------------------------------------------------

func TestSessionRepository_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)
		s := newTestSession(uuid.New().String())

		mock.ExpectExec(`update sessions`).
			WithArgs(
				s.ID,
				s.RefreshTokenHash,
				s.NotificationToken,
				s.Device.Hash,
				s.Device.Name,
				s.Device.Version,
				s.Device.Platform,
				s.ActiveIPAddress,
				s.ActiveAt,
				s.ExpiresAt,
				s.IsRevoked,
			).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(context.Background(), s)

		assert.NoError(t, err)
	})

	t.Run("not found — zero rows affected", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)

		mock.ExpectExec(`update sessions`).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Update(context.Background(), newTestSession(uuid.New().String()))

		assert.ErrorIs(t, err, domain.ErrSessionNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)
		dbErr := errors.New("connection refused")

		mock.ExpectExec(`update sessions`).
			WithArgs(
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
				sqlmock.AnyArg(), sqlmock.AnyArg(),
			).
			WillReturnError(dbErr)

		err := repo.Update(context.Background(), newTestSession(uuid.New().String()))

		assert.ErrorIs(t, err, dbErr)
	})
}

// ------------------------------------------------------------------------------------------------
// --	TerminateAllByUserID
// ------------------------------------------------------------------------------------------------

func TestSessionRepository_TerminateAllByUserID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)
		userID := uuid.New().String()

		mock.ExpectExec(`update sessions`).
			WithArgs(userID).
			WillReturnResult(sqlmock.NewResult(0, 3)) // 3 sessions terminated

		err := repo.TerminateAllByUserID(context.Background(), userID)

		assert.NoError(t, err)
	})

	t.Run("no sessions to terminate — zero rows affected", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)

		mock.ExpectExec(`update sessions`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.TerminateAllByUserID(context.Background(), uuid.New().String())

		assert.ErrorIs(t, err, domain.ErrAlreadyCleanSessions)
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)
		dbErr := errors.New("connection refused")

		mock.ExpectExec(`update sessions`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(dbErr)

		err := repo.TerminateAllByUserID(context.Background(), uuid.New().String())

		assert.ErrorIs(t, err, dbErr)
	})
}

// ------------------------------------------------------------------------------------------------
// --	TerminateBySessionID
// ------------------------------------------------------------------------------------------------

func TestSessionRepository_TerminateBySessionID(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)
		sessionID := uuid.New().String()

		mock.ExpectExec(`update sessions`).
			WithArgs(sessionID).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.TerminateBySessionID(context.Background(), sessionID)

		assert.NoError(t, err)
	})

	t.Run("not found — zero rows affected", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)

		mock.ExpectExec(`update sessions`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.TerminateBySessionID(context.Background(), uuid.New().String())

		assert.ErrorIs(t, err, domain.ErrSessionNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock := newSessionRepoMock(t)
		dbErr := errors.New("connection refused")

		mock.ExpectExec(`update sessions`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(dbErr)

		err := repo.TerminateBySessionID(context.Background(), uuid.New().String())

		assert.ErrorIs(t, err, dbErr)
	})
}
