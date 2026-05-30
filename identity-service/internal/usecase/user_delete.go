package usecase

import (
	"context"
	"errors"
	"fmt"
	"identity-service/internal/domain"
)

type UserDeleteInput struct {
	UserID   int
	Password string
}

// UserDelete coordinates the complete wiping of an identity profile.
type UserDelete struct {
	userRepo    domain.UserRepo
	sessionRepo domain.SessionRepo
	pwdHasher   domain.PasswordHasher
}

func NewUserDelete(
	userRepo domain.UserRepo,
	sessionRepo domain.SessionRepo,
	pwdHasher domain.PasswordHasher,
) *UserDelete {
	return &UserDelete{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		pwdHasher:   pwdHasher,
	}
}

// Execute performs security re-authentication and removes user presence.
func (ud *UserDelete) Execute(ctx context.Context, input UserDeleteInput) error {
	// 1. Fetch user data
	user, err := ud.userRepo.FindByID(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to look up account: %w", err)
	}
	if user == nil {
		return errors.New("user account not found")
	}

	// 2. Compare passwords
	match, err := ud.pwdHasher.Compare(user.PasswordHash, input.Password)
	if err != nil {
		return fmt.Errorf("failed to compare passwords: %w", err)
	}
	if !match {
		return errors.New("incorrect password")
	}

	// 3. Clear all active tracking sessions (Log out of all devices)
	err = ud.sessionRepo.DeleteAllByUserID(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to clear associated sessions: %w", err)
	}

	// 4. Delete the primary user profile entity
	err = ud.userRepo.Delete(ctx, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to finalize user erasure: %w", err)
	}

	return nil
}
