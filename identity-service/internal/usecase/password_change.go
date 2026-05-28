package usecase

import (
	"context"
	"errors"
	"fmt"
	"identity-service/internal/domain"
)

// 1. Determine Input
// 1.1. ChangePasswordInput defines the data required to change user password
type ChangePasswordInput struct {
	UserID          int
	CurrentPassword string
	NewPassword     string
}

// 2. Determine the dependencies
// 2.1. ChangePassword coordinates domain layer to change a password
type ChangePassword struct {
	userRepo  domain.UserRepo
	pwdHasher domain.PasswordHasher
}

// 2.2. NewChangePassword is a constructor that handles dependency injection
func NewChangePassword(userRepo domain.UserRepo, pwdHasher domain.PasswordHasher) *ChangePassword {
	return &ChangePassword{
		userRepo:  userRepo,
		pwdHasher: pwdHasher,
	}
}

// 3. Execute runs the actual step-by-step password change business flow
func (cp *ChangePassword) Execute(ctx context.Context, input ChangePasswordInput) error {
	// 1. Validate input basic constraints
	if input.UserID == 0 || input.CurrentPassword == "" || input.NewPassword == "" {
		return errors.New("required fields were not filled")
	}

	// 2. Find a user by phone
	user, err := cp.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to find a user: %w", err)
	}
	if user == nil {
		return errors.New("user has not been found")
	}

	// 3. Compare passwords + Prevent the user from reusing their exact same password
	match, err := cp.pwdHasher.Compare(user.PasswordHash, input.CurrentPassword)
	if err != nil {
		return fmt.Errorf("failed to compare passwords: %w", err)
	}
	if !match {
		return errors.New("incorrect password")
	}
	if input.CurrentPassword == input.NewPassword {
		return errors.New("new password cannot be the same as your current password")
	}

	// 4. Generate hash for the new password
	hashedPassword, err := cp.pwdHasher.Hash(input.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to process password: %w", err)
	}

	// 5. Change user password
	user.PasswordHash = hashedPassword
	err = cp.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save new password: %w", err)
	}

	return nil
}
