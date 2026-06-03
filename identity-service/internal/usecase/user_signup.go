package usecase

import (
	"context"
	"errors"
	"fmt"
	"identity-service/internal/domain"
)

// This use case was separated by 2 parts: request (send sms code) and confirm (verify sms code + register a new user)

// Part 1:
// Request: send otp code via sms

// 1. Determine the input
type SignUpRequestInput struct {
	Phone string
}

// 2. Determine the dependencies
type SignUpRequest struct {
	userRepo    domain.UserRepository
	smsProvider domain.SMSProvider
	otpGen      domain.OTPGenerator
	otpRepo     domain.OTPRepository
}

func NewSignUpRequest(
	userRepo domain.UserRepository,
	smsProvider domain.SMSProvider,
	otpGen domain.OTPGenerator,
	otpRepo domain.OTPRepository,
) *SignUpRequest {
	return &SignUpRequest{
		userRepo:    userRepo,
		smsProvider: smsProvider,
		otpGen:      otpGen,
		otpRepo:     otpRepo,
	}
}

// 3. Business flow of user registration (part 1: send an sms code to the user)
func (uc *SignUpRequest) Execute(ctx context.Context, input SignUpRequestInput) error {
	// 1. Verify that the user actually exists by phone number
	user, err := uc.userRepo.FindByPhone(ctx, input.Phone)
	if err != nil {
		return fmt.Errorf("failed to look up account: %w", err)
	}
	if user != nil {
		return errors.New("that phone number is already taken")
	}

	// 2. Generate a secure 6-digit numeric string
	code, err := uc.otpGen.Generate(6)
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}

	// 3. Persist the OTP with an expiration
	err = uc.otpRepo.Save(ctx, input.Phone, code)
	if err != nil {
		return fmt.Errorf("failed to process request: %w", err)
	}

	// 4. Send the SMS
	err = uc.smsProvider.SendOTP(ctx, input.Phone, code)
	if err != nil {
		return fmt.Errorf("failed to dispatch text message: %w", err)
	}

	return nil
}

// Part 2:
// Confirm: verify sms code and register a new user

// 1. Determine the input and the output
type SignUpConfirmInput struct {
	Username string
	Phone    string
	Code     string
	Password string
}

// 2. Determine the dependencies
type SignUpConfirm struct {
	userRepo  domain.UserRepository
	otpRepo   domain.OTPRepository
	pwdHasher domain.PasswordHasher
}

func NewSignUpConfirm(
	userRepo domain.UserRepository,
	otpRepo domain.OTPRepository,
	hasher domain.PasswordHasher,
) *SignUpConfirm {
	return &SignUpConfirm{
		userRepo:  userRepo,
		otpRepo:   otpRepo,
		pwdHasher: hasher,
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
	hashedPassword, err := uc.pwdHasher.Hash(input.Password)
	if err != nil {
		return fmt.Errorf("failed to process password: %w", err)
	}

	// 5. Register a new user
	userID := uc.userRepo.NewUUID()
	user, err := domain.NewUser(userID, input.Username, input.Phone, hashedPassword)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	err = uc.userRepo.Create(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to save user: %w", err)
	}

	return nil
}
