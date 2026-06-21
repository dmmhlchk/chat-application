package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"identity-service/internal/application/usecase"
	"identity-service/internal/domain"
)

// ___ Mocks _________________________________________________________________

type mockUserRepository struct{ mock.Mock }

func (m *mockUserRepository) FindByUserID(ctx context.Context, userID string) (*domain.User, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	args := m.Called(ctx, phone)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	args := m.Called(ctx, username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepository) ExistsByPhoneOrUsername(ctx context.Context, phone, username string) (bool, error) {
	args := m.Called(ctx, phone, username)
	return args.Bool(0), args.Error(1)
}

func (m *mockUserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *mockUserRepository) Update(ctx context.Context, user *domain.User) error {
	return m.Called(ctx, user).Error(0)
}

func (m *mockUserRepository) Delete(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

type mockPasswordHasher struct{ mock.Mock }

func (m *mockPasswordHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *mockPasswordHasher) Compare(hashedPassword, password string) (bool, error) {
	args := m.Called(hashedPassword, password)
	return args.Bool(0), args.Error(1)
}

type mockUUIDProvider struct{ mock.Mock }

func (m *mockUUIDProvider) Generate() string {
	return m.Called().String(0)
}

// ___ Helpers _________________________________________________________________

func newSignUpConfirmUC(
	uuidProvider *mockUUIDProvider,
	userRepo *mockUserRepository,
	otpRepo *mockOTPCacheRepository,
	hasher *mockPasswordHasher,
) *usecase.SignUpConfirm {
	return usecase.NewSignUpConfirm(uuidProvider, userRepo, otpRepo, hasher)
}

func validSignUpConfirmInput() usecase.SignUpConfirmInput {
	return usecase.SignUpConfirmInput{
		Username: "johndoe",
		Phone:    "+996700000000",
		Code:     "123456",
		Password: "S3cur3P@ss!",
	}
}

// ___ Tests _________________________________________________________________

func TestSignUpConfirm_Success(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockUUIDProvider{}
	userRepo := &mockUserRepository{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	input := validSignUpConfirmInput()

	otpRepo.On("Verify", ctx, input.Phone, input.Code).Return(true, nil)
	userRepo.On("ExistsByPhoneOrUsername", ctx, input.Phone, input.Username).Return(false, nil)
	hasher.On("Hash", input.Password).Return("$2a$hashed", nil)
	uuidProv.On("Generate").Return("new-uuid-1234")
	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil)

	uc := newSignUpConfirmUC(uuidProv, userRepo, otpRepo, hasher)
	err := uc.Execute(ctx, input)

	assert.NoError(t, err)
	userRepo.AssertExpectations(t)
	otpRepo.AssertExpectations(t)
	hasher.AssertExpectations(t)
	uuidProv.AssertExpectations(t)
}

func TestSignUpConfirm_MissingRequiredFields(t *testing.T) {
	cases := []struct {
		name  string
		input usecase.SignUpConfirmInput
	}{
		{
			name:  "empty username",
			input: usecase.SignUpConfirmInput{Phone: "+996700000000", Code: "123456", Password: "pass"},
		},
		{
			name:  "empty phone",
			input: usecase.SignUpConfirmInput{Username: "johndoe", Code: "123456", Password: "pass"},
		},
		{
			name:  "empty password",
			input: usecase.SignUpConfirmInput{Username: "johndoe", Phone: "+996700000000", Code: "123456"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uuidProv := &mockUUIDProvider{}
			userRepo := &mockUserRepository{}
			otpRepo := &mockOTPCacheRepository{}
			hasher := &mockPasswordHasher{}

			uc := newSignUpConfirmUC(uuidProv, userRepo, otpRepo, hasher)
			err := uc.Execute(context.Background(), tc.input)

			assert.EqualError(t, err, "required fields were not filled")
			otpRepo.AssertNotCalled(t, "Verify", mock.Anything, mock.Anything, mock.Anything)
		})
	}
}

func TestSignUpConfirm_InvalidOTPCode(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockUUIDProvider{}
	userRepo := &mockUserRepository{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	input := validSignUpConfirmInput()
	otpRepo.On("Verify", ctx, input.Phone, input.Code).Return(false, nil)

	uc := newSignUpConfirmUC(uuidProv, userRepo, otpRepo, hasher)
	err := uc.Execute(ctx, input)

	assert.EqualError(t, err, "invalid or expired verification code")
	userRepo.AssertNotCalled(t, "ExistsByPhoneOrUsername", mock.Anything, mock.Anything, mock.Anything)
}

func TestSignUpConfirm_OTPVerifyError(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockUUIDProvider{}
	userRepo := &mockUserRepository{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	input := validSignUpConfirmInput()
	otpRepo.On("Verify", ctx, input.Phone, input.Code).Return(false, errors.New("redis timeout"))

	uc := newSignUpConfirmUC(uuidProv, userRepo, otpRepo, hasher)
	err := uc.Execute(ctx, input)

	assert.EqualError(t, err, "invalid or expired verification code")
}

func TestSignUpConfirm_UserAlreadyExists(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockUUIDProvider{}
	userRepo := &mockUserRepository{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	input := validSignUpConfirmInput()
	otpRepo.On("Verify", ctx, input.Phone, input.Code).Return(true, nil)
	userRepo.On("ExistsByPhoneOrUsername", ctx, input.Phone, input.Username).Return(true, nil)

	uc := newSignUpConfirmUC(uuidProv, userRepo, otpRepo, hasher)
	err := uc.Execute(ctx, input)

	assert.EqualError(t, err, "phone number or username is already taken")
	hasher.AssertNotCalled(t, "Hash", mock.Anything)
}

func TestSignUpConfirm_ExistsCheckError(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockUUIDProvider{}
	userRepo := &mockUserRepository{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	input := validSignUpConfirmInput()
	otpRepo.On("Verify", ctx, input.Phone, input.Code).Return(true, nil)
	userRepo.On("ExistsByPhoneOrUsername", ctx, input.Phone, input.Username).Return(false, errors.New("db error"))

	uc := newSignUpConfirmUC(uuidProv, userRepo, otpRepo, hasher)
	err := uc.Execute(ctx, input)

	assert.ErrorContains(t, err, "failed to verify account uniqueness")
}

func TestSignUpConfirm_HashError(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockUUIDProvider{}
	userRepo := &mockUserRepository{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	input := validSignUpConfirmInput()
	otpRepo.On("Verify", ctx, input.Phone, input.Code).Return(true, nil)
	userRepo.On("ExistsByPhoneOrUsername", ctx, input.Phone, input.Username).Return(false, nil)
	hasher.On("Hash", input.Password).Return("", errors.New("bcrypt failure"))

	uc := newSignUpConfirmUC(uuidProv, userRepo, otpRepo, hasher)
	err := uc.Execute(ctx, input)

	assert.ErrorContains(t, err, "failed to process password")
	userRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestSignUpConfirm_CreateUserError(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockUUIDProvider{}
	userRepo := &mockUserRepository{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	input := validSignUpConfirmInput()
	otpRepo.On("Verify", ctx, input.Phone, input.Code).Return(true, nil)
	userRepo.On("ExistsByPhoneOrUsername", ctx, input.Phone, input.Username).Return(false, nil)
	hasher.On("Hash", input.Password).Return("$2a$hashed", nil)
	uuidProv.On("Generate").Return("new-uuid-1234")
	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(errors.New("db write error"))

	uc := newSignUpConfirmUC(uuidProv, userRepo, otpRepo, hasher)
	err := uc.Execute(ctx, input)

	assert.ErrorContains(t, err, "failed to save user")
}
