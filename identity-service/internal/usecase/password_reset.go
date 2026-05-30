package usecase

import (
	"context"
	"errors"
	"fmt"
	"identity-service/internal/domain"
)

// This use case was separated on 2 parts: request (send sms code) and confirm (verify sms code + reset password)

// Request: send otp code via sms

type PasswordResetRequestInput struct {
	Phone string
}

type PasswordResetRequest struct {
	userRepo     domain.UserRepo
	otpGenerator domain.OTPGenerator
	codeHandler  domain.CodeHandler
	smsProvider  domain.SMSProvider
}

func NewPasswordResetRequest(
	userRepo domain.UserRepo,
	otpGenerator domain.OTPGenerator,
	codeHandler domain.CodeHandler,
	smsProvider domain.SMSProvider,
) *PasswordResetRequest {
	return &PasswordResetRequest{
		userRepo:     userRepo,
		otpGenerator: otpGenerator,
		codeHandler:  codeHandler,
		smsProvider:  smsProvider,
	}
}

func (prr *PasswordResetRequest) Execute(ctx context.Context, input PasswordResetRequestInput) error {
	// 1. Verify that the user actually exists by phone number
	user, err := prr.userRepo.FindByPhone(ctx, input.Phone)
	if err != nil {
		return fmt.Errorf("failed to look up account: %w", err)
	}
	if user == nil {
		// Security tip: Return nil or a generic success to prevent phone scraping attacks,
		// or explicitly fail if your application UX demands showing an error.
		return errors.New("phone number not registered")
	}

	// 2. Generate a secure 6-digit numeric string
	code, err := prr.otpGenerator.Generate(6)
	if err != nil {
		return fmt.Errorf("failed to generate verification token: %w", err)
	}

	// 3. Persist the OTP with an expiration (e.g., 5 mins) inside Redis via the repository
	err = prr.codeHandler.Save(ctx, input.Phone, code)
	if err != nil {
		return fmt.Errorf("failed to process request: %w", err)
	}

	// 4. Send the SMS via your infrastructure adapter
	err = prr.smsProvider.SendOTP(ctx, input.Phone, code)
	if err != nil {
		return fmt.Errorf("failed to dispatch text message: %w", err)
	}

	return nil
}

// Confirm: verify otp code and reset password

type ResetConfirmInput struct {
	Phone       string
	Code        string
	NewPassword string
}

type PasswordResetConfirm struct {
	userRepo    domain.UserRepo
	codeHandler domain.CodeHandler
	pwdHasher   domain.PasswordHasher
}

func NewPasswordResetConfirm(
	userRepo domain.UserRepo,
	codeHandler domain.CodeHandler,
	pwdHasher domain.PasswordHasher,
) *PasswordResetConfirm {
	return &PasswordResetConfirm{
		userRepo:    userRepo,
		codeHandler: codeHandler,
		pwdHasher:   pwdHasher,
	}
}

func (uc *PasswordResetConfirm) Execute(ctx context.Context, input ResetConfirmInput) error {
	// 1. Validate password strength constraints
	if len(input.NewPassword) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	// 2. Ask the code repository to verify the code matches what we stored
	isValid, err := uc.codeHandler.Verify(ctx, input.Phone, input.Code)
	if err != nil || !isValid {
		return errors.New("invalid or expired verification code")
	}

	// 3. Code is correct! Find the target user profile
	user, err := uc.userRepo.FindByPhone(ctx, input.Phone)
	if err != nil || user == nil {
		return errors.New("user account no longer exists")
	}

	// 4. Hash the fresh password string
	hashedPassword, err := uc.pwdHasher.Hash(input.NewPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 5. Mutate the User Entity pointer and save it
	user.PasswordHash = hashedPassword
	err = uc.userRepo.Update(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// 6. Clean up the used verification code so it can't be replayed
	_ = uc.codeHandler.Delete(ctx, input.Phone)

	return nil
}
