package usecase

import (
	"context"
	"errors"
	"fmt"
	"identity-service/internal/domain"
)

// 1. Determine the input
type UserDeleteInput struct {
	UserID   string
	Password string
}

// 2. Determine the dependencies
type UserDelete struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	pwdHasher   domain.PasswordHasher
}

func NewUserDelete(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	pwdHasher domain.PasswordHasher,
) *UserDelete {
	return &UserDelete{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		pwdHasher:   pwdHasher,
	}
}

// 3. Business flow of deleting user data
func (uc *UserDelete) Execute(ctx context.Context, input UserDeleteInput) error {
	// 1. Fetch user data
	user, err := uc.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to look up account: %w", err)
	}
	if user == nil {
		return errors.New("user account not found")
	}

	// 2. Compare passwords
	match, err := uc.pwdHasher.Compare(user.PasswordHash, input.Password)
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
