package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"identity-service/internal/application/port"
)

// 1. Determine the input
type SignUpRequestInput struct {
	Phone string
}

// 2. Determine the dependencies
type SignUpRequest struct {
	userRepo  port.UserRepository
	otpSender port.OTPSender
	otpGen    port.OTPGenerator
	otpRepo   port.OTPCacheRepository
}

func NewSignUpRequest(
	userRepo port.UserRepository,
	otpSender port.OTPSender,
	otpGen port.OTPGenerator,
	otpRepo port.OTPCacheRepository,
) *SignUpRequest {
	return &SignUpRequest{
		userRepo:  userRepo,
		otpSender: otpSender,
		otpGen:    otpGen,
		otpRepo:   otpRepo,
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
	err = uc.otpRepo.Save(ctx, input.Phone, code, 1*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to process request: %w", err)
	}

	// 4. Send the SMS
	err = uc.otpSender.Send(ctx, input.Phone, code)
	if err != nil {
		return fmt.Errorf("failed to dispatch text message: %w", err)
	}

	return nil
}
