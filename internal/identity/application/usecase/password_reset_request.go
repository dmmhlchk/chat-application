package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"chat-app/internal/identity/application/generator"
	"chat-app/internal/identity/application/publisher"
	"chat-app/internal/identity/application/repository"
	"chat-app/internal/identity/domain"
)

// 1. Determine the input
type PasswordResetRequestInput struct {
	Phone string
}

// 2. Determine the dependencies
type PasswordResetRequest struct {
	userReader     repository.UserReader
	eventPublisher publisher.EventPublisher
	otpGen         generator.OTPGenerator
	otpRepo        repository.OTPCacheRepository
}

func NewPasswordResetRequest(
	userReader repository.UserReader,
	eventPublisher publisher.EventPublisher,
	otpGen generator.OTPGenerator,
	otpRepo repository.OTPCacheRepository,
) *PasswordResetRequest {
	return &PasswordResetRequest{
		userReader:     userReader,
		eventPublisher: eventPublisher,
		otpGen:         otpGen,
		otpRepo:        otpRepo,
	}
}

// 3. Busines flow of the reseting password (part 1: send an sms code to the user)
func (uc *PasswordResetRequest) Execute(ctx context.Context, input PasswordResetRequestInput) error {
	// 1. Verify that the user actually exists by phone number
	user, err := uc.userReader.FindByPhone(ctx, input.Phone)
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

	// 3. Persist the OTP with an expiration
	err = uc.otpRepo.Save(ctx, input.Phone, code, 1*time.Minute)
	if err != nil {
		return fmt.Errorf("failed to process request: %w", err)
	}

	// 4. Send the SMS
	evt := domain.OTPCreated{
		Phone: input.Phone,
		Code:  code,
	}

	err = uc.eventPublisher.PublishOTPCreated(ctx, evt)
	if err != nil {
		return fmt.Errorf("failed to dispatch text message: %w", err)
	}

	return nil
}
