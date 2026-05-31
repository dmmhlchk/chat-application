package usecase

import (
	"context"
	"errors"
	"fmt"
	"identity-service/internal/domain"
	"time"
)

// 1. Determine the input and the output
type SignInInput struct {
	Phone             string
	Password          string
	NotificationToken string
	Device            domain.Device
	IPAddress         string
}

type SignInOutput struct {
	UserId           int
	AccessToken      string
	RefreshTokenHash string
}

// 2. Determine the dependencies
type SignIn struct {
	userRepo    domain.UserRepo
	sessionRepo domain.SessionRepo
	pwdHasher   domain.PasswordHasher
	tokenGen    domain.TokenGenerator
}

func NewSignIn(
	userRepo domain.UserRepo,
	sessionRepo domain.SessionRepo,
	pwdHasher domain.PasswordHasher,
	tokenGen domain.TokenGenerator,
) *SignIn {
	return &SignIn{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		pwdHasher:   pwdHasher,
		tokenGen:    tokenGen,
	}
}

// 3. Business flow of user authentication
func (uc *SignIn) Execute(ctx context.Context, input SignInInput) (*SignInOutput, error) {
	// 1. Validate input basic constraints
	if input.Phone == "" || input.Password == "" {
		return nil, errors.New("required fields were not filled")
	}

	// 2. Find a user by phone
	user, err := uc.userRepo.FindByPhone(ctx, input.Phone)
	if err != nil {
		return nil, fmt.Errorf("failed to find a user by phone number: %w", err)
	}
	if user == nil {
		return nil, errors.New("invalid username or password")
	}

	// 3. Compare passwords
	match, err := uc.pwdHasher.Compare(user.PasswordHash, input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to compare passwords: %w", err)
	}
	if !match {
		return nil, errors.New("incorrect password")
	}

	// 4. Generate auth tokens
	expiration := 30 * 24 * time.Hour
	accessToken, refreshTokenHash, err := uc.tokenGen.GeneratePair(user.ID, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate authentication tokens: %w", err)
	}

	// 5. Register a new session
	newSession := domain.NewSession(
		user.ID,
		refreshTokenHash,
		input.NotificationToken,
		input.Device,
		input.IPAddress,
		expiration,
	)

	// 6. Save session to database
	err = uc.sessionRepo.Create(ctx, newSession)
	if err != nil {
		return nil, fmt.Errorf("failed to establish secure session: %w", err)
	}

	// 7. Return success
	return &SignInOutput{
		UserId:           user.ID,
		AccessToken:      accessToken,
		RefreshTokenHash: refreshTokenHash,
	}, nil
}
