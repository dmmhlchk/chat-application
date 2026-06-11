package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"identity-service/internal/application/port"
	"identity-service/internal/domain"

	"github.com/google/uuid"
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
	UserID       uuid.UUID
	AccessToken  string
	RefreshToken string
}

// 2. Determine the dependencies
type SignIn struct {
	uuidGen     port.UUIDGeneratod
	userRepo    port.UserRepository
	sessionRepo port.SessionRepository
	pwdHasher   port.PasswordHasher
	tokenGen    port.TokenGenerator
}

func NewSignIn(
	uuidGen port.UUIDGeneratod,
	userRepo port.UserRepository,
	sessionRepo port.SessionRepository,
	pwdHasher port.PasswordHasher,
	tokenGen port.TokenGenerator,
) *SignIn {
	return &SignIn{
		uuidGen:     uuidGen,
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
	sessionID := uc.uuidGen.NewUUID()
	var expiration time.Duration

	expiration = 15 * time.Minute
	accessToken, err := uc.tokenGen.GenerateToken(user.ID, sessionID, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access tokens: %w", err)
	}

	expiration = 30 * 24 * time.Hour
	refreshToken, err := uc.tokenGen.GenerateToken(user.ID, sessionID, expiration)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh tokens: %w", err)
	}

	// 5. Register a new session
	session, err := domain.NewSession(
		sessionID,
		user.ID,
		refreshToken,
		input.NotificationToken,
		input.Device,
		input.IPAddress,
		expiration,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

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
