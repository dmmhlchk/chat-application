package postgres_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"chat-app/internal/identity/domain"
	"chat-app/internal/identity/infra/persistence/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ___ Helpers _________________________________________________________________
func newUserRepoMock(t *testing.T) (*postgres.UserRepository, sqlmock.Sqlmock) {
	t.Helper()

	db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Close()
		assert.NoError(t, mock.ExpectationsWereMet())
	})

	return postgres.NewUserRepository(db).(*postgres.UserRepository), mock
}

func userColumns() []string {
	return []string{"id", "username", "phone", "password_hash"}
}

func mockUserRow(u *domain.User) []driver.Value {
	return []driver.Value{u.ID, u.Username, u.Phone, u.PasswordHash}
}

func newTestUser() *domain.User {
	return &domain.User{
		ID:           uuid.New().String(),
		Username:     "john_doe",
		Phone:        "+996700000001",
		PasswordHash: "$2a$10$hashedpassword",
	}
}

// ___ Tests _________________________________________________________________
func TestUserRepository_FindByUserID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		user := newTestUser()

		mock.ExpectQuery(`select .+ from users where id = \$1`).
			WithArgs(user.ID).
			WillReturnRows(
				sqlmock.NewRows(userColumns()).AddRow(mockUserRow(user)...),
			)

		got, err := repo.FindByUserID(context.Background(), user.ID)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, user.ID, got.ID)
		assert.Equal(t, user.Username, got.Username)
		assert.Equal(t, user.Phone, got.Phone)
		assert.Equal(t, user.PasswordHash, got.PasswordHash)
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)

		mock.ExpectQuery(`select .+ from users where id = \$1`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(sql.ErrNoRows)

		got, err := repo.FindByUserID(context.Background(), uuid.New().String())

		assert.Nil(t, got)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		dbErr := errors.New("connection refused")

		mock.ExpectQuery(`select .+ from users where id = \$1`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(dbErr)

		got, err := repo.FindByUserID(context.Background(), uuid.New().String())

		assert.Nil(t, got)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestUserRepository_FindByPhone(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		user := newTestUser()

		mock.ExpectQuery(`select .+ from users where phone = \$1`).
			WithArgs(user.Phone).
			WillReturnRows(
				sqlmock.NewRows(userColumns()).AddRow(mockUserRow(user)...),
			)

		got, err := repo.FindByPhone(context.Background(), user.Phone)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, user.Phone, got.Phone)
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)

		mock.ExpectQuery(`select .+ from users where phone = \$1`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(sql.ErrNoRows)

		got, err := repo.FindByPhone(context.Background(), "+996700000099")

		assert.Nil(t, got)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		dbErr := errors.New("timeout")

		mock.ExpectQuery(`select .+ from users where phone = \$1`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(dbErr)

		got, err := repo.FindByPhone(context.Background(), "+996700000099")

		assert.Nil(t, got)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestUserRepository_FindByUsername(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		user := newTestUser()

		mock.ExpectQuery(`select .+ from users where username = \$1`).
			WithArgs(user.Username).
			WillReturnRows(
				sqlmock.NewRows(userColumns()).AddRow(mockUserRow(user)...),
			)

		got, err := repo.FindByUsername(context.Background(), user.Username)

		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, user.Username, got.Username)
	})

	t.Run("not found", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)

		mock.ExpectQuery(`select .+ from users where username = \$1`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(sql.ErrNoRows)

		got, err := repo.FindByUsername(context.Background(), "ghost")

		assert.Nil(t, got)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		dbErr := errors.New("timeout")

		mock.ExpectQuery(`select .+ from users where username = \$1`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(dbErr)

		got, err := repo.FindByUsername(context.Background(), "ghost")

		assert.Nil(t, got)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestUserRepository_ExistsByPhoneOrUsername(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)

		mock.ExpectQuery(`select exists`).
			WithArgs("john_doe", "+996700000001").
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(true))

		exists, err := repo.ExistsByPhoneOrUsername(context.Background(), "+996700000001", "john_doe")

		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("does not exist", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)

		mock.ExpectQuery(`select exists`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnRows(sqlmock.NewRows([]string{"exists"}).AddRow(false))

		exists, err := repo.ExistsByPhoneOrUsername(context.Background(), "+996700000099", "nobody")

		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		dbErr := errors.New("connection refused")

		mock.ExpectQuery(`select exists`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(dbErr)

		exists, err := repo.ExistsByPhoneOrUsername(context.Background(), "+996700000001", "john_doe")

		assert.False(t, exists)
		assert.ErrorIs(t, err, dbErr)
	})
}

func TestUserRepository_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		user := newTestUser()

		mock.ExpectExec(`insert into users`).
			WithArgs(user.ID, user.Username, user.Phone, user.PasswordHash).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := repo.Create(context.Background(), user)

		assert.NoError(t, err)
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		dbErr := errors.New("duplicate key value")

		mock.ExpectExec(`insert into users`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(dbErr)

		err := repo.Create(context.Background(), newTestUser())

		assert.ErrorIs(t, err, dbErr)
	})
}

func TestUserRepository_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		user := newTestUser()

		mock.ExpectExec(`update users`).
			WithArgs(user.ID, user.Username, user.Phone, user.PasswordHash).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Update(context.Background(), user)

		assert.NoError(t, err)
	})

	t.Run("not found — zero rows affected", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)

		mock.ExpectExec(`update users`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0)) // 0 rows affected

		err := repo.Update(context.Background(), newTestUser())

		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		dbErr := errors.New("connection refused")

		mock.ExpectExec(`update users`).
			WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
			WillReturnError(dbErr)

		err := repo.Update(context.Background(), newTestUser())

		assert.ErrorIs(t, err, dbErr)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		id := uuid.New().String()

		mock.ExpectExec(`delete from users`).
			WithArgs(id).
			WillReturnResult(sqlmock.NewResult(0, 1))

		err := repo.Delete(context.Background(), id)

		assert.NoError(t, err)
	})

	t.Run("not found — zero rows affected", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)

		mock.ExpectExec(`delete from users`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(0, 0))

		err := repo.Delete(context.Background(), uuid.New().String())

		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})

	t.Run("db error", func(t *testing.T) {
		repo, mock := newUserRepoMock(t)
		dbErr := errors.New("connection refused")

		mock.ExpectExec(`delete from users`).
			WithArgs(sqlmock.AnyArg()).
			WillReturnError(dbErr)

		err := repo.Delete(context.Background(), uuid.New().String())

		assert.ErrorIs(t, err, dbErr)
	})
}
