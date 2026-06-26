package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"chat-app/service/identity/application/usecase"
	"chat-app/service/identity/domain"
)

// ___ Helpers _________________________________________________________________

func newSignUpRequestUC(
	userReader *mockUserReader,
	publisher *mockEventPublisher,
	otpGen *mockOTPGenerator,
	otpRepo *mockOTPCacheRepository,
) *usecase.SignUpRequest {
	return usecase.NewSignUpRequest(userReader, publisher, otpGen, otpRepo)
}

// ___ Tests _________________________________________________________________

func TestSignUpRequest_Success(t *testing.T) {
	ctx := context.Background()

	userReader := &mockUserReader{}
	publisher := &mockEventPublisher{}
	otpGen := &mockOTPGenerator{}
	otpRepo := &mockOTPCacheRepository{}

	userReader.On("FindByPhone", ctx, "+996700000000").Return(nil, nil)
	otpGen.On("Generate", 6).Return("123456", nil)
	otpRepo.On("Save", ctx, "+996700000000", "123456", 1*time.Minute).Return(nil)
	publisher.On("PublishOTPCreated", ctx, domain.OTPCreated{
		Phone: "+996700000000",
		Code:  "123456",
	}).Return(nil)

	uc := newSignUpRequestUC(userReader, publisher, otpGen, otpRepo)
	err := uc.Execute(ctx, usecase.SignUpRequestInput{Phone: "+996700000000"})

	assert.NoError(t, err)
	userReader.AssertExpectations(t)
	otpGen.AssertExpectations(t)
	otpRepo.AssertExpectations(t)
	publisher.AssertExpectations(t)
}

func TestSignUpRequest_PhoneAlreadyTaken(t *testing.T) {
	ctx := context.Background()

	userReader := &mockUserReader{}
	publisher := &mockEventPublisher{}
	otpGen := &mockOTPGenerator{}
	otpRepo := &mockOTPCacheRepository{}

	existingUser := &domain.User{ID: "some-uuid"}
	userReader.On("FindByPhone", ctx, "+996700000000").Return(existingUser, nil)

	uc := newSignUpRequestUC(userReader, publisher, otpGen, otpRepo)
	err := uc.Execute(ctx, usecase.SignUpRequestInput{Phone: "+996700000000"})

	assert.EqualError(t, err, "that phone number is already taken")
	// OTP should never be generated or persisted
	otpGen.AssertNotCalled(t, "Generate", mock.Anything)
	otpRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	publisher.AssertNotCalled(t, "PublishOTPCreated", mock.Anything, mock.Anything)
}

func TestSignUpRequest_FindByPhoneError(t *testing.T) {
	ctx := context.Background()

	userReader := &mockUserReader{}
	publisher := &mockEventPublisher{}
	otpGen := &mockOTPGenerator{}
	otpRepo := &mockOTPCacheRepository{}

	userReader.On("FindByPhone", ctx, "+996700000000").Return(nil, errors.New("db error"))

	uc := newSignUpRequestUC(userReader, publisher, otpGen, otpRepo)
	err := uc.Execute(ctx, usecase.SignUpRequestInput{Phone: "+996700000000"})

	assert.ErrorContains(t, err, "failed to look up account")
}

func TestSignUpRequest_OTPGenerationError(t *testing.T) {
	ctx := context.Background()

	userReader := &mockUserReader{}
	publisher := &mockEventPublisher{}
	otpGen := &mockOTPGenerator{}
	otpRepo := &mockOTPCacheRepository{}

	userReader.On("FindByPhone", ctx, "+996700000000").Return(nil, nil)
	otpGen.On("Generate", 6).Return("", errors.New("entropy exhausted"))

	uc := newSignUpRequestUC(userReader, publisher, otpGen, otpRepo)
	err := uc.Execute(ctx, usecase.SignUpRequestInput{Phone: "+996700000000"})

	assert.ErrorContains(t, err, "failed to generate verification token")
	otpRepo.AssertNotCalled(t, "Save", mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestSignUpRequest_OTPSaveError(t *testing.T) {
	ctx := context.Background()

	userReader := &mockUserReader{}
	publisher := &mockEventPublisher{}
	otpGen := &mockOTPGenerator{}
	otpRepo := &mockOTPCacheRepository{}

	userReader.On("FindByPhone", ctx, "+996700000000").Return(nil, nil)
	otpGen.On("Generate", 6).Return("123456", nil)
	otpRepo.On("Save", ctx, "+996700000000", "123456", 1*time.Minute).Return(errors.New("redis down"))

	uc := newSignUpRequestUC(userReader, publisher, otpGen, otpRepo)
	err := uc.Execute(ctx, usecase.SignUpRequestInput{Phone: "+996700000000"})

	assert.ErrorContains(t, err, "failed to process request")
	publisher.AssertNotCalled(t, "PublishOTPCreated", mock.Anything, mock.Anything)
}

func TestSignUpRequest_PublishEventError(t *testing.T) {
	ctx := context.Background()

	userReader := &mockUserReader{}
	publisher := &mockEventPublisher{}
	otpGen := &mockOTPGenerator{}
	otpRepo := &mockOTPCacheRepository{}

	userReader.On("FindByPhone", ctx, "+996700000000").Return(nil, nil)
	otpGen.On("Generate", 6).Return("123456", nil)
	otpRepo.On("Save", ctx, "+996700000000", "123456", 1*time.Minute).Return(nil)
	publisher.On("PublishOTPCreated", ctx, domain.OTPCreated{
		Phone: "+996700000000",
		Code:  "123456",
	}).Return(errors.New("broker unavailable"))

	uc := newSignUpRequestUC(userReader, publisher, otpGen, otpRepo)
	err := uc.Execute(ctx, usecase.SignUpRequestInput{Phone: "+996700000000"})

	assert.ErrorContains(t, err, "failed to dispatch text message")
}
