package usecase

import (
	"context"
	"errors"
	"fmt"
	"identity-service/internal/domain"
)

// This use case was separated by 2 parts: request (send sms code) and confirm (verify sms code + reset the password)

// Part 1:
// Request: send otp code via sms

// 1. Determine the input
type PasswordResetRequestInput struct {
	Phone string
}

// 2. Determine the dependencies
type PasswordResetRequest struct {
	userRepo    domain.UserRepository
	smsProvider domain.SMSProvider
	otpGen      domain.OTPGenerator
	otpRepo     domain.OTPRepository
}

func NewPasswordResetRequest(
	userRepo domain.UserRepository,
	smsProvider domain.SMSProvider,
	otpGen domain.OTPGenerator,
	otpRepo domain.OTPRepository,
) *PasswordResetRequest {
	return &PasswordResetRequest{
		userRepo:    userRepo,
		smsProvider: smsProvider,
		otpGen:      otpGen,
		otpRepo:     otpRepo,
	}
}

// 3. Busines flow of the reseting password (part 1: send an sms code to the user)
func (uc *PasswordResetRequest) Execute(ctx context.Context, input PasswordResetRequestInput) error {
	// 1. Verify that the user actually exists by phone number
	user, err := uc.userRepo.FindByPhone(ctx, input.Phone)
	if err != nil {
		return fmt.Errorf("failed to look up account: %w", err)
	}
	if user == nil {
		return errors.New("phone number not registered")
	}

	// 2. Generate a secure 6-digit numeric string
	code, err := uc.otpGen.Generate(6)
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}

	// 3. Persist the OTP with an expiration (e.g., 5 mins) inside Redis via the repository
	err = uc.otpRepo.Save(ctx, input.Phone, code)
	if err != nil {
		return fmt.Errorf("failed to process request: %w", err)
	}

	// 4. Send the SMS via your infrastructure adapter
	err = uc.smsProvider.SendOTP(ctx, input.Phone, code)
	if err != nil {
		return fmt.Errorf("failed to dispatch text message: %w", err)
	}

	return nil
}

// Part 2:
// Confirm: verify sms code and reset the password

// 1. Determine the input
type ResetConfirmInput struct {
	Phone       string
	Code        string
	NewPassword string
}

// 2. Determine the dependencies
type PasswordResetConfirm struct {
	userRepo    domain.UserRepository
	sessionRepo domain.SessionRepository
	otpRepo     domain.OTPRepository
	pwdHasher   domain.PasswordHasher
}

func NewPasswordResetConfirm(
	userRepo domain.UserRepository,
	sessionRepo domain.SessionRepository,
	otpRepo domain.OTPRepository,
	pwdHasher domain.PasswordHasher,
) *PasswordResetConfirm {
	return &PasswordResetConfirm{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		otpRepo:     otpRepo,
		pwdHasher:   pwdHasher,
	}
}

// 3. Busines flow of the reseting password (part 2: verify sms code + reset the password)
func (uc *PasswordResetConfirm) Execute(ctx context.Context, input ResetConfirmInput) error {
	// 1. Validate password strength constraints
	if len(input.NewPassword) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// 2. Verify the code matches what we stored
	isValid, err := uc.otpRepo.Verify(ctx, input.Phone, input.Code)
	if err != nil || !isValid {
		return errors.New("invalid or expired verification code")
	}

	// 3. Find the target user profile
	user, err := uc.userRepo.FindByPhone(ctx, input.Phone)
	if err != nil || user == nil {
		return errors.New("user account no longer exists")
	}

	// 4. Revoke all active sessions
	err = uc.sessionRepo.TerminateAll(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed termainta all sessions: %w", err)
	}

	// 5. Hash the fresh password string
	hashedPassword, err := uc.pwdHasher.Hash(input.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 6. Mutate the User Entity pointer and save it
	user.PasswordHash = hashedPassword
	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 7. Clean up the used verification code so it can't be replayed
	_ = uc.otpRepo.Delete(ctx, input.Phone)

	return nil
}
