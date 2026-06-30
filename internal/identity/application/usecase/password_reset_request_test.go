package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"chat-app/internal/identity/application/usecase"
	"chat-app/internal/identity/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ___ Tests _________________________________________________________________
func TestPasswordResetRequest_Success(t *testing.T) {
	ctx := context.Background()

	userReader := &mockUserReader{}
	publisher := &mockEventPublisher{}
	otpGen := &mockOTPGenerator{}
	otpRepo := &mockOTPCacheRepository{}

	user := fakeUser()

	userReader.On("FindByPhone", ctx, "+996700000000").
		Return(user, nil)

	otpGen.On("Generate", 6).
		Return("654321", nil)

	otpRepo.On("Save", ctx, "+996700000000", "654321", 1*time.Minute).
		Return(nil)

	publisher.On("PublishOTPCreated", ctx, domain.OTPCreated{
		Phone: "+996700000000",
		Code:  "654321",
	}).
		Return(nil)

	uc := usecase.NewPasswordResetRequest(userReader, publisher, otpGen, otpRepo)
	err := uc.Execute(ctx, usecase.PasswordResetRequestInput{Phone: "+996700000000"})

	assert.NoError(t, err)
	userReader.AssertExpectations(t)
	otpGen.AssertExpectations(t)
	otpRepo.AssertExpectations(t)
	publisher.AssertExpectations(t)
}

func TestPasswordResetRequest_PhoneNotRegistered(t *testing.T) {
	ctx := context.Background()

	userReader := &mockUserReader{}
	publisher := &mockEventPublisher{}
	otpGen := &mockOTPGenerator{}
	otpRepo := &mockOTPCacheRepository{}

	// Unlike SignUpRequest, reset requires the user to exist
	userReader.On("FindByPhone", ctx, "+996700000000").
		Return(nil, nil)

	uc := usecase.NewPasswordResetRequest(userReader, publisher, otpGen, otpRepo)
	err := uc.Execute(ctx, usecase.PasswordResetRequestInput{Phone: "+996700000000"})

	assert.EqualError(t, err, "phone number not registered")
	otpGen.AssertNotCalled(t, "Generate", mock.Anything)
	otpRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	publisher.AssertNotCalled(t, "PublishOTPCreated", mock.Anything, mock.Anything)
}

func TestPasswordResetRequest_FindByPhoneError(t *testing.T) {
	ctx := context.Background()

	userReader := &mockUserReader{}
	publisher := &mockEventPublisher{}
	otpGen := &mockOTPGenerator{}
	otpRepo := &mockOTPCacheRepository{}

	userReader.On("FindByPhone", ctx, "+996700000000").
		Return(nil, errors.New("db error"))

	uc := usecase.NewPasswordResetRequest(userReader, publisher, otpGen, otpRepo)
	err := uc.Execute(ctx, usecase.PasswordResetRequestInput{Phone: "+996700000000"})

	assert.ErrorContains(t, err, "failed to look up account")
}

func TestPasswordResetRequest_OTPGenerationError(t *testing.T) {
	ctx := context.Background()

	userReader := &mockUserReader{}
	publisher := &mockEventPublisher{}
	otpGen := &mockOTPGenerator{}
	otpRepo := &mockOTPCacheRepository{}

	user := fakeUser()

	userReader.On("FindByPhone", ctx, "+996700000000").
		Return(user, nil)

	otpGen.On("Generate", 6).
		Return("", errors.New("entropy exhausted"))

	uc := usecase.NewPasswordResetRequest(userReader, publisher, otpGen, otpRepo)
	err := uc.Execute(ctx, usecase.PasswordResetRequestInput{Phone: "+996700000000"})

	assert.ErrorContains(t, err, "failed to generate verification token")
	otpRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestPasswordResetRequest_OTPSaveError(t *testing.T) {
	ctx := context.Background()

	userReader := &mockUserReader{}
	publisher := &mockEventPublisher{}
	otpGen := &mockOTPGenerator{}
	otpRepo := &mockOTPCacheRepository{}

	user := fakeUser()

	userReader.On("FindByPhone", ctx, "+996700000000").
		Return(user, nil)

	otpGen.On("Generate", 6).
		Return("654321", nil)

	otpRepo.On("Save", ctx, "+996700000000", "654321", 1*time.Minute).
		Return(errors.New("redis down"))

	uc := usecase.NewPasswordResetRequest(userReader, publisher, otpGen, otpRepo)
	err := uc.Execute(ctx, usecase.PasswordResetRequestInput{Phone: "+996700000000"})

	assert.ErrorContains(t, err, "failed to process request")
	publisher.AssertNotCalled(t, "PublishOTPCreated", mock.Anything, mock.Anything)
}

func TestPasswordResetRequest_PublishEventError(t *testing.T) {
	ctx := context.Background()

	userReader := &mockUserReader{}
	publisher := &mockEventPublisher{}
	otpGen := &mockOTPGenerator{}
	otpRepo := &mockOTPCacheRepository{}

	user := fakeUser()

	userReader.On("FindByPhone", ctx, "+996700000000").
		Return(user, nil)

	otpGen.On("Generate", 6).
		Return("654321", nil)

	otpRepo.On("Save", ctx, "+996700000000", "654321", 1*time.Minute).
		Return(nil)

	publisher.On("PublishOTPCreated", ctx, domain.OTPCreated{
		Phone: "+996700000000",
		Code:  "654321",
	}).
		Return(errors.New("broker unavailable"))

	uc := usecase.NewPasswordResetRequest(userReader, publisher, otpGen, otpRepo)
	err := uc.Execute(ctx, usecase.PasswordResetRequestInput{Phone: "+996700000000"})

	assert.ErrorContains(t, err, "failed to dispatch text message")
}
