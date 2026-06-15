package usecase

import (
	"context"
	"errors"
	"fmt"
	"identity-service/internal/application/port"

	"github.com/google/uuid"
)

// 1. Determine the input
type UserDeleteInput struct {
	UserID   uuid.UUID
	Password string
}

// 2. Determine the dependencies
type UserDelete struct {
	userRepo       port.UserRepository
	passwordHasher port.PasswordHasher
}

func NewUserDelete(
	userRepo port.UserRepository,
	passwordHasher port.PasswordHasher,
) *UserDelete {
	return &UserDelete{
		userRepo:       userRepo,
		passwordHasher: passwordHasher,
	}
}

// 3. Business flow of deleting user data
func (uc *UserDelete) Execute(ctx context.Context, input UserDeleteInput) error {
	// 1. Fetch user data
	user, err := uc.userRepo.FindByUserID(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to look up account: %w", err)
	}
	if user == nil {
		return errors.New("user account not found")
	}

	// 2. Compare passwords
	match, err := uc.passwordHasher.Compare(user.PasswordHash, input.Password)
	if err != nil {
		return fmt.Errorf("failed to compare passwords: %w", err)
	}
	if !match {
		return errors.New("incorrect password")
	}

	// 3. Delete the primary user profile entity
	err = uc.userRepo.Delete(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to finalize user erasure: %w", err)
	}

	return nil
}
