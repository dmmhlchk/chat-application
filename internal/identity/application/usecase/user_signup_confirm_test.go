package usecase_test

import (
	"context"
	"errors"
	"testing"

	"chat-app/internal/identity/application/usecase"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// ___ Helpers _________________________________________________________________
func newSignUpConfirmUC(
	uuidProvider *mockIDGenerator,
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

	uuidProv := &mockIDGenerator{}
	userRepo := &mockUserRepository{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	input := validSignUpConfirmInput()

	otpRepo.On("Verify", ctx, input.Phone, input.Code).
		Return(true, nil)

	userRepo.On("ExistsByPhoneOrUsername", ctx, input.Phone, input.Username).
		Return(false, nil)

	hasher.On("Hash", input.Password).
		Return("$2a$hashed", nil)

	uuidProv.On("Generate").
		Return("new-uuid-1234")

	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).
		Return(nil)

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
			uuidProv := &mockIDGenerator{}
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

	uuidProv := &mockIDGenerator{}
	userRepo := &mockUserRepository{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	input := validSignUpConfirmInput()

	otpRepo.On("Verify", ctx, input.Phone, input.Code).
		Return(false, nil)

	uc := newSignUpConfirmUC(uuidProv, userRepo, otpRepo, hasher)
	err := uc.Execute(ctx, input)

	assert.EqualError(t, err, "invalid or expired verification code")
	userRepo.AssertNotCalled(t, "ExistsByPhoneOrUsername", mock.Anything, mock.Anything, mock.Anything)
}

func TestSignUpConfirm_OTPVerifyError(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockIDGenerator{}
	userRepo := &mockUserRepository{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	input := validSignUpConfirmInput()

	otpRepo.On("Verify", ctx, input.Phone, input.Code).
		Return(false, errors.New("redis timeout"))

	uc := newSignUpConfirmUC(uuidProv, userRepo, otpRepo, hasher)
	err := uc.Execute(ctx, input)

	assert.EqualError(t, err, "invalid or expired verification code")
}

func TestSignUpConfirm_UserAlreadyExists(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockIDGenerator{}
	userRepo := &mockUserRepository{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	input := validSignUpConfirmInput()

	otpRepo.On("Verify", ctx, input.Phone, input.Code).
		Return(true, nil)

	userRepo.On("ExistsByPhoneOrUsername", ctx, input.Phone, input.Username).
		Return(true, nil)

	uc := newSignUpConfirmUC(uuidProv, userRepo, otpRepo, hasher)
	err := uc.Execute(ctx, input)

	assert.EqualError(t, err, "phone number or username is already taken")
	hasher.AssertNotCalled(t, "Hash", mock.Anything)
}

func TestSignUpConfirm_ExistsCheckError(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockIDGenerator{}
	userRepo := &mockUserRepository{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	input := validSignUpConfirmInput()

	otpRepo.On("Verify", ctx, input.Phone, input.Code).
		Return(true, nil)

	userRepo.On("ExistsByPhoneOrUsername", ctx, input.Phone, input.Username).
		Return(false, errors.New("db error"))

	uc := newSignUpConfirmUC(uuidProv, userRepo, otpRepo, hasher)
	err := uc.Execute(ctx, input)

	assert.ErrorContains(t, err, "failed to verify account uniqueness")
}

func TestSignUpConfirm_HashError(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockIDGenerator{}
	userRepo := &mockUserRepository{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	input := validSignUpConfirmInput()

	otpRepo.On("Verify", ctx, input.Phone, input.Code).
		Return(true, nil)

	userRepo.On("ExistsByPhoneOrUsername", ctx, input.Phone, input.Username).
		Return(false, nil)

	hasher.On("Hash", input.Password).
		Return("", errors.New("bcrypt failure"))

	uc := newSignUpConfirmUC(uuidProv, userRepo, otpRepo, hasher)
	err := uc.Execute(ctx, input)

	assert.ErrorContains(t, err, "failed to process password")
	userRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestSignUpConfirm_CreateUserError(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockIDGenerator{}
	userRepo := &mockUserRepository{}
	otpRepo := &mockOTPCacheRepository{}
	hasher := &mockPasswordHasher{}

	input := validSignUpConfirmInput()

	otpRepo.On("Verify", ctx, input.Phone, input.Code).
		Return(true, nil)

	userRepo.On("ExistsByPhoneOrUsername", ctx, input.Phone, input.Username).
		Return(false, nil)

	hasher.On("Hash", input.Password).
		Return("$2a$hashed", nil)

	uuidProv.On("Generate").
		Return("new-uuid-1234")

	userRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).
		Return(errors.New("db write error"))

	uc := newSignUpConfirmUC(uuidProv, userRepo, otpRepo, hasher)
	err := uc.Execute(ctx, input)

	assert.ErrorContains(t, err, "failed to save user")
}
