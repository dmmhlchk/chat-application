package usecase

import (
	"context"
	"errors"
	"fmt"

	"identity-service/internal/application/port"
)

// 1. Determine the input
type ResetConfirmInput struct {
	Phone       string
	Code        string
	NewPassword string
}

// 2. Determine the dependencies
type PasswordResetConfirm struct {
	userRepo       port.UserRepository
	sessionWriter  port.SessionWriter
	otpRepo        port.OTPCacheRepository
	passwordHasher port.PasswordHasher
}

func NewPasswordResetConfirm(
	userRepo port.UserRepository,
	sessionWriter port.SessionWriter,
	otpRepo port.OTPCacheRepository,
	passwordHasher port.PasswordHasher,
) *PasswordResetConfirm {
	return &PasswordResetConfirm{
		userRepo:       userRepo,
		sessionWriter:  sessionWriter,
		otpRepo:        otpRepo,
		passwordHasher: passwordHasher,
	}
}

// 3. Busines flow of the reseting password (part 2: verify sms code + reset the password)
func (uc *PasswordResetConfirm) Execute(ctx context.Context, input ResetConfirmInput) error {
	// 1. Validate password strength constraints
	if len(input.NewPassword) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// 2. Verify the code matches what we stored
	isValid, err := uc.otpRepo.Verify(ctx, input.Phone, input.Code)
	if err != nil || !isValid {
		return errors.New("invalid or expired verification code")
	}

	// 3. Find the target user profile
	user, err := uc.userRepo.FindByPhone(ctx, input.Phone)
	if err != nil || user == nil {
		return errors.New("user account no longer exists")
	}

	// 4. Revoke all active sessions
	err = uc.sessionWriter.TerminateAllByUserID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed terminate all sessions: %w", err)
	}

	// 5. Hash the fresh password string
	hashedPassword, err := uc.passwordHasher.Hash(input.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 6. Change user password
	user.PasswordHash = hashedPassword
	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 7. Clean up the used verification code so it can't be replayed
	_ = uc.otpRepo.Delete(ctx, input.Phone)

	return nil
}
