package usecase

import (
	"context"
	"errors"
	"fmt"
	"identity-service/internal/domain"
	"time"
)

// 1. Determine Input and Output data
// 1.1. SignInInput defines the data required to identify the user
type SignInInput struct {
	Phone             string
	Password          string
	NotificationToken string
	Device            domain.Device
	IPAddress         string
}

// 1.2. SignInOutput defines what data we return to the delivery layer upon success
type SignInOutput struct {
	UserId           int
	AccessToken      string
	RefreshTokenHash string
}

// 2. Determine the dependencies
// 2.1. SignIn coordinates domain layer to sign in a user
type SignIn struct {
	userRepo    domain.UserRepo
	sessionRepo domain.SessionRepo
	pwdHasher   domain.PasswordHasher
	tokenGen    domain.TokenGenerator
}

// 2.2. NewSignIn is a constructor that handles dependency injection
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

// 3. Execute runs the actual step-by-step sign in business flow
func (si *SignIn) Execute(ctx context.Context, input SignInInput) (*SignInOutput, error) {
	// 1. Validate input basic constraints
	if input.Phone == "" || input.Password == "" {
		return nil, errors.New("required fields were not filled")
	}

	// 2. Find a user by phone
	user, err := si.userRepo.FindByPhone(ctx, input.Phone)
	if err != nil {
		return nil, fmt.Errorf("failed to find a user by phone number: %w", err)
	}
	if user == nil {
		return nil, errors.New("invalid username or password")
	}

	// 3. Compare passwords
	match, err := si.pwdHasher.Compare(user.PasswordHash, input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to compare passwords: %w", err)
	}
	if !match {
		return nil, errors.New("incorrect password")
	}

	// 4. Generate auth tokens
	expiration := 30 * 24 * time.Hour
	accessToken, refreshTokenHash, err := si.tokenGen.GeneratePair(user.ID, expiration)
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
	err = si.sessionRepo.Create(ctx, newSession)
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
