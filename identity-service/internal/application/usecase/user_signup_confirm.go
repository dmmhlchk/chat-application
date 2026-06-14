package usecase

import (
	"context"
	"errors"
	"fmt"

	"identity-service/internal/application/port"
	"identity-service/internal/domain"
)

// 1. Determine the input and the output
type SignUpConfirmInput struct {
	Username string
	Phone    string
	Code     string
	Password string
}

// 2. Determine the dependencies
type SignUpConfirm struct {
	uuidProvider   port.UUIDProvider
	userRepo       port.UserRepository
	otpRepo        port.OTPCacheRepository
	passwordHasher port.PasswordHasher
}

func NewSignUpConfirm(
	uuidProvider port.UUIDProvider,
	userRepo port.UserRepository,
	otpRepo port.OTPCacheRepository,
	passwordHasher port.PasswordHasher,
) *SignUpConfirm {
	return &SignUpConfirm{
		uuidProvider:   uuidProvider,
		userRepo:       userRepo,
		otpRepo:        otpRepo,
		passwordHasher: passwordHasher,
	}
}

// 3. Business flow of user registration (part 2: verify sms code + register a new user)
func (uc *SignUpConfirm) Execute(ctx context.Context, input SignUpConfirmInput) error {
	// 1. Validate input basic constraints
	if input.Username == "" || input.Phone == "" || input.Password == "" {
		return errors.New("required fields were not filled")
	}

	// 2. Verify the code matches what we stored
	isValid, err := uc.otpRepo.Verify(ctx, input.Phone, input.Code)
	if err != nil || !isValid {
		return errors.New("invalid or expired verification code")
	}

	// 3. Check the domain repository to see if this user already exists
	exists, err := uc.userRepo.ExistsByPhoneOrUsername(ctx, input.Phone, input.Username)
	if err != nil {
		return fmt.Errorf("failed to verify account uniqueness: %w", err)
	}
	if exists {
		return errors.New("phone number or username is already registered")
	}

	// 4. Secure the password using our domain's abstract PasswordHasher interface
	hashedPassword, err := uc.passwordHasher.Hash(input.Password)
	if err != nil {
		return fmt.Errorf("failed to process password: %w", err)
	}

	// 5. Register a new user
	userID := uc.uuidProvider.Generate()
	user, err := domain.NewUser(
		userID,
		input.Username,
		input.Phone,
		hashedPassword,
	)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	err = uc.userRepo.Create(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}
