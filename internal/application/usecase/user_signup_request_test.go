package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"chat-application/internal/application/usecase"
	"chat-application/internal/domain"
)

// ___ Mocks _________________________________________________________________

type mockUserReader struct{ mock.Mock }

func (m *mockUserReader) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserReader) FindByUserID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserReader) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserReader) ExistsByPhoneOrUsername(ctx context.Context, phone, username string) (bool, error) {
	args := m.Called(ctx, phone, username)
	return args.Bool(0), args.Error(1)
}

type mockEventPublisher struct{ mock.Mock }

func (m *mockEventPublisher) PublishUserCreated(ctx context.Context, evt domain.UserCreated) error {
	return m.Called(ctx, evt).Error(0)
}

func (m *mockEventPublisher) PublishOTPCreated(ctx context.Context, evt domain.OTPCreated) error {
	return m.Called(ctx, evt).Error(0)
}

type mockOTPGenerator struct{ mock.Mock }

func (m *mockOTPGenerator) Generate(length int) (string, error) {
	args := m.Called(length)
	return args.String(0), args.Error(1)
}

type mockOTPCacheRepository struct{ mock.Mock }

func (m *mockOTPCacheRepository) Save(ctx context.Context, phone, code string, ttl time.Duration) error {
	return m.Called(ctx, phone, code, ttl).Error(0)
}

func (m *mockOTPCacheRepository) Verify(ctx context.Context, phone, code string) (bool, error) {
	args := m.Called(ctx, phone, code)
	return args.Bool(0), args.Error(1)
}

func (m *mockOTPCacheRepository) Delete(ctx context.Context, phone string) error {
	return m.Called(ctx, phone).Error(0)
}

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
