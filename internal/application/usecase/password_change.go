package usecase

import (
	"context"
	"errors"
	"fmt"

	"internal/application/port"
)

// 1. Determine the input
type ChangePasswordInput struct {
	UserID          string
	CurrentPassword string
	NewPassword     string
}

// 2. Determine the dependencies
type ChangePassword struct {
	userRepo       port.UserRepository
	sessionWriter  port.SessionWriter
	passwordHasher port.PasswordHasher
}

func NewChangePassword(
	userRepo port.UserRepository,
	sessionWriter port.SessionWriter,
	passwordHasher port.PasswordHasher,
) *ChangePassword {
	return &ChangePassword{
		userRepo:       userRepo,
		sessionWriter:  sessionWriter,
		passwordHasher: passwordHasher,
	}
}

// 3. Business flow of changing password
func (uc *ChangePassword) Execute(ctx context.Context, input ChangePasswordInput) error {
	// 1. Validate input basic constraints
	if input.UserID == "" || input.CurrentPassword == "" || input.NewPassword == "" {
		return errors.New("required fields were not filled")
	}

	// 2. Find a user by phone
	user, err := uc.userRepo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to find a user: %w", err)
	}
	if user == nil {
		return errors.New("user has not been found")
	}

	// 3. Compare passwords + Prevent the user from reusing their exact same password
	match, err := uc.passwordHasher.Compare(user.PasswordHash, input.CurrentPassword)
	if err != nil {
		return fmt.Errorf("failed to compare passwords: %w", err)
	}
	if !match {
		return errors.New("incorrect password")
	}
	if input.CurrentPassword == input.NewPassword {
		return errors.New("new password cannot be the same as your current password")
	}

	// 4. Revoke all active sessions
	err = uc.sessionWriter.TerminateAllByUserID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed terminate all sessions: %w", err)
	}

	// 5. Generate hash for the new password
	hashedPassword, err := uc.passwordHasher.Hash(input.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to process password: %w", err)
	}

	// 6. Change user password
	user.PasswordHash = hashedPassword
	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save new password: %w", err)
	}

	return nil
}
