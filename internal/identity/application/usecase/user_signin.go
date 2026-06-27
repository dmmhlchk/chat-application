package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"chat-app/internal/identity/application/crypto"
	"chat-app/internal/identity/application/generator"
	"chat-app/internal/identity/application/repository"
	"chat-app/internal/identity/domain"
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
	UserID       string
	AccessToken  string
	RefreshToken string
}

// 2. Determine the dependencies
type SignIn struct {
	idGen          generator.IDGenerator
	userRepo       repository.UserRepository
	sessionRepo    repository.SessionRepository
	passwordHasher crypto.PasswordHasher
	tokenGen       generator.TokenGenerator
}

func NewSignIn(
	idGen generator.IDGenerator,
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
	passwordHasher crypto.PasswordHasher,
	tokenGen generator.TokenGenerator,
) *SignIn {
	return &SignIn{
		idGen:          idGen,
		userRepo:       userRepo,
		sessionRepo:    sessionRepo,
		passwordHasher: passwordHasher,
		tokenGen:       tokenGen,
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
	match, err := uc.passwordHasher.Compare(user.PasswordHash, input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to compare passwords: %w", err)
	}
	if !match {
		return nil, errors.New("incorrect password")
	}

	// 4. Generate auth tokens
	sessionID := uc.idGen.Generate()
	var expiration time.Duration

	expiration = 15 * time.Minute
	accessToken, err := uc.tokenGen.Generate(user.ID, sessionID, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access tokens: %w", err)
	}

	expiration = 30 * 24 * time.Hour
	refreshToken, err := uc.tokenGen.Generate(user.ID, sessionID, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh tokens: %w", err)
	}

	// 5. Register a new session
	session := domain.NewSession(
		sessionID,
		user.ID,
		refreshToken,
		input.NotificationToken,
		input.Device,
		input.IPAddress,
		expiration,
	)

	// 6. Save session to database
	err = uc.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to establish secure session: %w", err)
	}

	// 7. Return success
	return &SignInOutput{
		UserID:       user.ID,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
