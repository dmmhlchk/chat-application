package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"chat-application/internal/application/usecase"
	"chat-application/internal/domain"
)

// ___ Mock _________________________________________________________________

type mockSessionWriter struct{ mock.Mock }

func (m *mockSessionWriter) Create(ctx context.Context, session *domain.Session) error {
	return m.Called(ctx, session).Error(0)
}

func (m *mockSessionWriter) Update(ctx context.Context, session *domain.Session) error {
	return m.Called(ctx, session).Error(0)
}

func (m *mockSessionWriter) TerminateAllByUserID(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockSessionWriter) TerminateBySessionID(ctx context.Context, sessionID string) error {
	return m.Called(ctx, sessionID).Error(0)
}

// ___ Tests _________________________________________________________________
func TestChangePassword_Success(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()
	userRepo.On("FindByUserID", ctx, "user-uuid-001").Return(user, nil)
	hasher.On("Compare", user.PasswordHash, "OldP@ss1!").Return(true, nil)
	sessionWriter.On("TerminateAllByUserID", ctx, user.ID).Return(nil)
	hasher.On("Hash", "NewP@ss2!").Return("$2a$new-hash", nil)
	userRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	uc := usecase.NewChangePassword(userRepo, sessionWriter, hasher)
	err := uc.Execute(ctx, usecase.ChangePasswordInput{
		UserID:          "user-uuid-001",
		CurrentPassword: "OldP@ss1!",
		NewPassword:     "NewP@ss2!",
	})

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	sessionWriter.AssertExpectations(t)
	hasher.AssertExpectations(t)
}

func TestChangePassword_MissingRequiredFields(t *testing.T) {
	cases := []struct {
		name  string
		input usecase.ChangePasswordInput
	}{
		{
			name:  "empty userID",
			input: usecase.ChangePasswordInput{CurrentPassword: "old", NewPassword: "new"},
		},
		{
			name:  "empty currentPassword",
			input: usecase.ChangePasswordInput{UserID: "user-uuid-001", NewPassword: "new"},
		},
		{
			name:  "empty newPassword",
			input: usecase.ChangePasswordInput{UserID: "user-uuid-001", CurrentPassword: "old"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			userRepo := &mockUserRepository{}
			sessionWriter := &mockSessionWriter{}
			hasher := &mockPasswordHasher{}

			uc := usecase.NewChangePassword(userRepo, sessionWriter, hasher)
			err := uc.Execute(context.Background(), tc.input)

			assert.EqualError(t, err, "required fields were not filled")
			userRepo.AssertNotCalled(t, "FindByUserID", mock.Anything, mock.Anything)
		})
	}
}

func TestChangePassword_UserNotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	hasher := &mockPasswordHasher{}

	userRepo.On("FindByUserID", ctx, "user-uuid-001").Return(nil, nil)

	uc := usecase.NewChangePassword(userRepo, sessionWriter, hasher)
	err := uc.Execute(ctx, usecase.ChangePasswordInput{
		UserID:          "user-uuid-001",
		CurrentPassword: "OldP@ss1!",
		NewPassword:     "NewP@ss2!",
	})

	assert.EqualError(t, err, "user has not been found")
	hasher.AssertNotCalled(t, "Compare", mock.Anything, mock.Anything)
}

func TestChangePassword_FindByUserIDError(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	hasher := &mockPasswordHasher{}

	userRepo.On("FindByUserID", ctx, "user-uuid-001").Return(nil, errors.New("db error"))

	uc := usecase.NewChangePassword(userRepo, sessionWriter, hasher)
	err := uc.Execute(ctx, usecase.ChangePasswordInput{
		UserID:          "user-uuid-001",
		CurrentPassword: "OldP@ss1!",
		NewPassword:     "NewP@ss2!",
	})

	assert.ErrorContains(t, err, "failed to find a user")
}

func TestChangePassword_WrongCurrentPassword(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()
	userRepo.On("FindByUserID", ctx, "user-uuid-001").Return(user, nil)
	hasher.On("Compare", user.PasswordHash, "WrongP@ss!").Return(false, nil)

	uc := usecase.NewChangePassword(userRepo, sessionWriter, hasher)
	err := uc.Execute(ctx, usecase.ChangePasswordInput{
		UserID:          "user-uuid-001",
		CurrentPassword: "WrongP@ss!",
		NewPassword:     "NewP@ss2!",
	})

	assert.EqualError(t, err, "incorrect password")
	sessionWriter.AssertNotCalled(t, "TerminateAllByUserID", mock.Anything, mock.Anything)
}

func TestChangePassword_CompareError(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()
	userRepo.On("FindByUserID", ctx, "user-uuid-001").Return(user, nil)
	hasher.On("Compare", user.PasswordHash, "OldP@ss1!").Return(false, errors.New("bcrypt error"))

	uc := usecase.NewChangePassword(userRepo, sessionWriter, hasher)
	err := uc.Execute(ctx, usecase.ChangePasswordInput{
		UserID:          "user-uuid-001",
		CurrentPassword: "OldP@ss1!",
		NewPassword:     "NewP@ss2!",
	})

	assert.ErrorContains(t, err, "failed to compare passwords")
}

func TestChangePassword_SamePassword(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()
	userRepo.On("FindByUserID", ctx, "user-uuid-001").Return(user, nil)
	// Compare succeeds — password is correct, but new == current
	hasher.On("Compare", user.PasswordHash, "SameP@ss1!").Return(true, nil)

	uc := usecase.NewChangePassword(userRepo, sessionWriter, hasher)
	err := uc.Execute(ctx, usecase.ChangePasswordInput{
		UserID:          "user-uuid-001",
		CurrentPassword: "SameP@ss1!",
		NewPassword:     "SameP@ss1!",
	})

	assert.EqualError(t, err, "new password cannot be the same as your current password")
	sessionWriter.AssertNotCalled(t, "TerminateAllByUserID", mock.Anything, mock.Anything)
}

func TestChangePassword_TerminateSessionsError(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()
	userRepo.On("FindByUserID", ctx, "user-uuid-001").Return(user, nil)
	hasher.On("Compare", user.PasswordHash, "OldP@ss1!").Return(true, nil)
	sessionWriter.On("TerminateAllByUserID", ctx, user.ID).Return(errors.New("db error"))

	uc := usecase.NewChangePassword(userRepo, sessionWriter, hasher)
	err := uc.Execute(ctx, usecase.ChangePasswordInput{
		UserID:          "user-uuid-001",
		CurrentPassword: "OldP@ss1!",
		NewPassword:     "NewP@ss2!",
	})

	assert.ErrorContains(t, err, "failed terminate all sessions")
	hasher.AssertNotCalled(t, "Hash", mock.Anything)
}

func TestChangePassword_HashNewPasswordError(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()
	userRepo.On("FindByUserID", ctx, "user-uuid-001").Return(user, nil)
	hasher.On("Compare", user.PasswordHash, "OldP@ss1!").Return(true, nil)
	sessionWriter.On("TerminateAllByUserID", ctx, user.ID).Return(nil)
	hasher.On("Hash", "NewP@ss2!").Return("", errors.New("bcrypt error"))

	uc := usecase.NewChangePassword(userRepo, sessionWriter, hasher)
	err := uc.Execute(ctx, usecase.ChangePasswordInput{
		UserID:          "user-uuid-001",
		CurrentPassword: "OldP@ss1!",
		NewPassword:     "NewP@ss2!",
	})

	assert.ErrorContains(t, err, "failed to process password")
	userRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestChangePassword_UpdateUserError(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()
	userRepo.On("FindByUserID", ctx, "user-uuid-001").Return(user, nil)
	hasher.On("Compare", user.PasswordHash, "OldP@ss1!").Return(true, nil)
	sessionWriter.On("TerminateAllByUserID", ctx, user.ID).Return(nil)
	hasher.On("Hash", "NewP@ss2!").Return("$2a$new-hash", nil)
	userRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Return(errors.New("db write error"))

	uc := usecase.NewChangePassword(userRepo, sessionWriter, hasher)
	err := uc.Execute(ctx, usecase.ChangePasswordInput{
		UserID:          "user-uuid-001",
		CurrentPassword: "OldP@ss1!",
		NewPassword:     "NewP@ss2!",
	})

	assert.ErrorContains(t, err, "failed to save new password")
}
