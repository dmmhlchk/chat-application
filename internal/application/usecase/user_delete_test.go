package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"chat-application/internal/application/usecase"
)

// ___ Tests _________________________________________________________________

func TestUserDelete_Success(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()
	userRepo.On("FindByUserID", ctx, "user-uuid-001").Return(user, nil)
	hasher.On("Compare", user.PasswordHash, "S3cur3P@ss!").Return(true, nil)
	userRepo.On("Delete", ctx, "user-uuid-001").Return(nil)

	uc := usecase.NewUserDelete(userRepo, hasher)
	err := uc.Execute(ctx, usecase.UserDeleteInput{
		UserID:   "user-uuid-001",
		Password: "S3cur3P@ss!",
	})

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	hasher.AssertExpectations(t)
}

func TestUserDelete_UserNotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	hasher := &mockPasswordHasher{}

	userRepo.On("FindByUserID", ctx, "user-uuid-001").Return(nil, nil)

	uc := usecase.NewUserDelete(userRepo, hasher)
	err := uc.Execute(ctx, usecase.UserDeleteInput{
		UserID:   "user-uuid-001",
		Password: "S3cur3P@ss!",
	})

	assert.EqualError(t, err, "user account not found")
	hasher.AssertNotCalled(t, "Compare", mock.Anything, mock.Anything)
	userRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestUserDelete_FindByUserIDError(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	hasher := &mockPasswordHasher{}

	userRepo.On("FindByUserID", ctx, "user-uuid-001").Return(nil, errors.New("db error"))

	uc := usecase.NewUserDelete(userRepo, hasher)
	err := uc.Execute(ctx, usecase.UserDeleteInput{
		UserID:   "user-uuid-001",
		Password: "S3cur3P@ss!",
	})

	assert.ErrorContains(t, err, "failed to look up account")
	hasher.AssertNotCalled(t, "Compare", mock.Anything, mock.Anything)
}

func TestUserDelete_WrongPassword(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()
	userRepo.On("FindByUserID", ctx, "user-uuid-001").Return(user, nil)
	hasher.On("Compare", user.PasswordHash, "wrongpassword").Return(false, nil)

	uc := usecase.NewUserDelete(userRepo, hasher)
	err := uc.Execute(ctx, usecase.UserDeleteInput{
		UserID:   "user-uuid-001",
		Password: "wrongpassword",
	})

	assert.EqualError(t, err, "incorrect password")
	userRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestUserDelete_ComparePasswordError(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()
	userRepo.On("FindByUserID", ctx, "user-uuid-001").Return(user, nil)
	hasher.On("Compare", user.PasswordHash, "S3cur3P@ss!").Return(false, errors.New("bcrypt error"))

	uc := usecase.NewUserDelete(userRepo, hasher)
	err := uc.Execute(ctx, usecase.UserDeleteInput{
		UserID:   "user-uuid-001",
		Password: "S3cur3P@ss!",
	})

	assert.ErrorContains(t, err, "failed to compare passwords")
	userRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestUserDelete_DeleteError(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()
	userRepo.On("FindByUserID", ctx, "user-uuid-001").Return(user, nil)
	hasher.On("Compare", user.PasswordHash, "S3cur3P@ss!").Return(true, nil)
	userRepo.On("Delete", ctx, "user-uuid-001").Return(errors.New("db write error"))

	uc := usecase.NewUserDelete(userRepo, hasher)
	err := uc.Execute(ctx, usecase.UserDeleteInput{
		UserID:   "user-uuid-001",
		Password: "S3cur3P@ss!",
	})

	assert.ErrorContains(t, err, "failed to finalize user erasure")
}
