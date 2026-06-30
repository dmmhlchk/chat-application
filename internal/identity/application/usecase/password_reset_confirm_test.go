package usecase_test

import (
	"context"
	"errors"
	"testing"

	"chat-app/internal/identity/application/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ___ Tests _________________________________________________________________
func TestPasswordResetConfirm_Success(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()

	otpRepo.On("Verify", ctx, "+996700000000", "654321").
		Return(true, nil)

	userRepo.On("FindByPhone", ctx, "+996700000000").
		Return(user, nil)

	sessionWriter.On("TerminateAllByUserID", ctx, user.ID).
		Return(nil)

	hasher.On("Hash", "NewP@ss2!").
		Return("$2a$new-hash", nil)

	userRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).
		Return(nil)

	otpRepo.On("Delete", ctx, "+996700000000").
		Return(nil)

	uc := usecase.NewPasswordResetConfirm(userRepo, sessionWriter, otpRepo, hasher)
	err := uc.Execute(ctx, usecase.PasswordResetConfirmInput{
		Phone:       "+996700000000",
		Code:        "654321",
		NewPassword: "NewP@ss2!",
	})

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	sessionWriter.AssertExpectations(t)
	otpRepo.AssertExpectations(t)
	hasher.AssertExpectations(t)
}

func TestPasswordResetConfirm_PasswordTooShort(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	uc := usecase.NewPasswordResetConfirm(userRepo, sessionWriter, otpRepo, hasher)
	err := uc.Execute(ctx, usecase.PasswordResetConfirmInput{
		Phone:       "+996700000000",
		Code:        "654321",
		NewPassword: "short",
	})

	assert.EqualError(t, err, "password must be at least 8 characters long")
	otpRepo.AssertNotCalled(t, "Verify", mock.Anything, mock.Anything, mock.Anything)
}

func TestPasswordResetConfirm_PasswordExactly8Chars(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()

	otpRepo.On("Verify", ctx, "+996700000000", "654321").
		Return(true, nil)

	userRepo.On("FindByPhone", ctx, "+996700000000").
		Return(user, nil)

	sessionWriter.On("TerminateAllByUserID", ctx, user.ID).
		Return(nil)

	hasher.On("Hash", "Exact8Ch").
		Return("$2a$hash", nil)

	userRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).
		Return(nil)

	otpRepo.On("Delete", ctx, "+996700000000").
		Return(nil)

	uc := usecase.NewPasswordResetConfirm(userRepo, sessionWriter, otpRepo, hasher)
	err := uc.Execute(ctx, usecase.PasswordResetConfirmInput{
		Phone:       "+996700000000",
		Code:        "654321",
		NewPassword: "Exact8Ch",
	})

	assert.NoError(t, err)
}

func TestPasswordResetConfirm_InvalidOTPCode(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	otpRepo.On("Verify", ctx, "+996700000000", "000000").
		Return(false, nil)

	uc := usecase.NewPasswordResetConfirm(userRepo, sessionWriter, otpRepo, hasher)
	err := uc.Execute(ctx, usecase.PasswordResetConfirmInput{
		Phone:       "+996700000000",
		Code:        "000000",
		NewPassword: "NewP@ss2!",
	})

	assert.EqualError(t, err, "invalid or expired verification code")
	userRepo.AssertNotCalled(t, "FindByPhone", mock.Anything, mock.Anything)
}

func TestPasswordResetConfirm_OTPVerifyError(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	otpRepo.On("Verify", ctx, "+996700000000", "654321").
		Return(false, errors.New("redis timeout"))

	uc := usecase.NewPasswordResetConfirm(userRepo, sessionWriter, otpRepo, hasher)
	err := uc.Execute(ctx, usecase.PasswordResetConfirmInput{
		Phone:       "+996700000000",
		Code:        "654321",
		NewPassword: "NewP@ss2!",
	})

	assert.EqualError(t, err, "invalid or expired verification code")
}

func TestPasswordResetConfirm_UserNotFound(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	otpRepo.On("Verify", ctx, "+996700000000", "654321").
		Return(true, nil)

	userRepo.On("FindByPhone", ctx, "+996700000000").
		Return(nil, nil)

	uc := usecase.NewPasswordResetConfirm(userRepo, sessionWriter, otpRepo, hasher)
	err := uc.Execute(ctx, usecase.PasswordResetConfirmInput{
		Phone:       "+996700000000",
		Code:        "654321",
		NewPassword: "NewP@ss2!",
	})

	assert.EqualError(t, err, "user account no longer exists")
	sessionWriter.AssertNotCalled(t, "TerminateAllByUserID", mock.Anything, mock.Anything)
}

func TestPasswordResetConfirm_FindByPhoneError(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	otpRepo.On("Verify", ctx, "+996700000000", "654321").
		Return(true, nil)

	userRepo.On("FindByPhone", ctx, "+996700000000").
		Return(nil, errors.New("db error"))

	uc := usecase.NewPasswordResetConfirm(userRepo, sessionWriter, otpRepo, hasher)
	err := uc.Execute(ctx, usecase.PasswordResetConfirmInput{
		Phone:       "+996700000000",
		Code:        "654321",
		NewPassword: "NewP@ss2!",
	})

	assert.EqualError(t, err, "user account no longer exists")
}

func TestPasswordResetConfirm_TerminateSessionsError(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()

	otpRepo.On("Verify", ctx, "+996700000000", "654321").
		Return(true, nil)

	userRepo.On("FindByPhone", ctx, "+996700000000").
		Return(user, nil)

	sessionWriter.On("TerminateAllByUserID", ctx, user.ID).
		Return(errors.New("db error"))

	uc := usecase.NewPasswordResetConfirm(userRepo, sessionWriter, otpRepo, hasher)
	err := uc.Execute(ctx, usecase.PasswordResetConfirmInput{
		Phone:       "+996700000000",
		Code:        "654321",
		NewPassword: "NewP@ss2!",
	})

	assert.ErrorContains(t, err, "failed terminate all sessions")
	hasher.AssertNotCalled(t, "Hash", mock.Anything)
}

func TestPasswordResetConfirm_HashError(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()

	otpRepo.On("Verify", ctx, "+996700000000", "654321").
		Return(true, nil)

	userRepo.On("FindByPhone", ctx, "+996700000000").
		Return(user, nil)

	sessionWriter.On("TerminateAllByUserID", ctx, user.ID).
		Return(nil)

	hasher.On("Hash", "NewP@ss2!").
		Return("", errors.New("bcrypt error"))

	uc := usecase.NewPasswordResetConfirm(userRepo, sessionWriter, otpRepo, hasher)
	err := uc.Execute(ctx, usecase.PasswordResetConfirmInput{
		Phone:       "+996700000000",
		Code:        "654321",
		NewPassword: "NewP@ss2!",
	})

	assert.ErrorContains(t, err, "failed to hash password")
	userRepo.AssertNotCalled(t, "Update", mock.Anything, mock.Anything)
}

func TestPasswordResetConfirm_UpdateUserError(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()

	otpRepo.On("Verify", ctx, "+996700000000", "654321").
		Return(true, nil)

	userRepo.On("FindByPhone", ctx, "+996700000000").
		Return(user, nil)

	sessionWriter.On("TerminateAllByUserID", ctx, user.ID).
		Return(nil)

	hasher.On("Hash", "NewP@ss2!").
		Return("$2a$new-hash", nil)

	userRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).
		Return(errors.New("db write error"))

	uc := usecase.NewPasswordResetConfirm(userRepo, sessionWriter, otpRepo, hasher)
	err := uc.Execute(ctx, usecase.PasswordResetConfirmInput{
		Phone:       "+996700000000",
		Code:        "654321",
		NewPassword: "NewP@ss2!",
	})

	assert.ErrorContains(t, err, "failed to update password")
	otpRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
}

func TestPasswordResetConfirm_OTPDeleteFailureIsIgnored(t *testing.T) {
	ctx := context.Background()

	userRepo := &mockUserRepository{}
	sessionWriter := &mockSessionWriter{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	user := fakeUser()

	otpRepo.On("Verify", ctx, "+996700000000", "654321").
		Return(true, nil)

	userRepo.On("FindByPhone", ctx, "+996700000000").
		Return(user, nil)

	sessionWriter.On("TerminateAllByUserID", ctx, user.ID).
		Return(nil)

	hasher.On("Hash", "NewP@ss2!").
		Return("$2a$new-hash", nil)

	userRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).
		Return(nil)

	// Delete fails but the use case swallows the error with `_ =`
	otpRepo.On("Delete", ctx, "+996700000000").
		Return(errors.New("redis error"))

	uc := usecase.NewPasswordResetConfirm(userRepo, sessionWriter, otpRepo, hasher)
	err := uc.Execute(ctx, usecase.PasswordResetConfirmInput{
		Phone:       "+996700000000",
		Code:        "654321",
		NewPassword: "NewP@ss2!",
	})

	// Should still succeed — Delete error is intentionally ignored
	assert.NoError(t, err)
}
