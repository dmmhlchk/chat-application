package usecase_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"internal/application/usecase"
	"internal/domain"
)

// ___ Mocks _________________________________________________________________

type mockSessionRepository struct{ mock.Mock }

func (m *mockSessionRepository) FindAllByUserID(ctx context.Context, userID string) ([]domain.Session, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Session), args.Error(1)
}

func (m *mockSessionRepository) FindBySessionID(ctx context.Context, sessionID string) (*domain.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

func (m *mockSessionRepository) Create(ctx context.Context, session *domain.Session) error {
	return m.Called(ctx, session).Error(0)
}

func (m *mockSessionRepository) Update(ctx context.Context, session *domain.Session) error {
	return m.Called(ctx, session).Error(0)
}

func (m *mockSessionRepository) TerminateAllByUserID(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockSessionRepository) TerminateBySessionID(ctx context.Context, sessionID string) error {
	return m.Called(ctx, sessionID).Error(0)
}

type mockTokenGenerator struct{ mock.Mock }

func (m *mockTokenGenerator) GenerateToken(userID, sessionID string, ttl time.Duration) (string, error) {
	args := m.Called(userID, sessionID, ttl)
	return args.String(0), args.Error(1)
}

func (m *mockTokenGenerator) ValidateToken(token string) (string, string, error) {
	args := m.Called(token)
	return args.String(0), args.String(1), args.Error(2)
}

// ___ Helpers _________________________________________________________________

func newSignInUC(
	uuidProv *mockUUIDProvider,
	userRepo *mockUserRepository,
	sessionRepo *mockSessionRepository,
	hasher *mockPasswordHasher,
	tokenGen *mockTokenGenerator,
) *usecase.SignIn {
	return usecase.NewSignIn(uuidProv, userRepo, sessionRepo, hasher, tokenGen)
}

func validSignInInput() usecase.SignInInput {
	return usecase.SignInInput{
		Phone:             "+996700000000",
		Password:          "S3cur3P@ss!",
		NotificationToken: "fcm-token-xyz",
		Device: domain.Device{
			Hash:     "c191537f-df38-417b-ad14-00b6d10117b1",
			Name:     "iPhone 15",
			Version:  "17",
			Platform: domain.PlatformIOS,
		},
		IPAddress: "192.168.1.1",
	}
}

func fakeUser() *domain.User {
	return &domain.User{
		ID:           "user-uuid-001",
		Username:     "johndoe",
		Phone:        "+996700000000",
		PasswordHash: "$2a$hashed",
	}
}

// ___ Tests _________________________________________________________________

func TestSignIn_Success(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockUUIDProvider{}
	userRepo := &mockUserRepository{}
	sessionRepo := &mockSessionRepository{}
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}

	input := validSignInInput()
	user := fakeUser()

	userRepo.On("FindByPhone", ctx, input.Phone).Return(user, nil)
	hasher.On("Compare", user.PasswordHash, input.Password).Return(true, nil)
	uuidProv.On("Generate").Return("session-uuid-001")
	tokenGen.On("GenerateToken", user.ID, "session-uuid-001", mock.Anything).
		Return("access-token", nil).Once()
	tokenGen.On("GenerateToken", user.ID, "session-uuid-001", mock.Anything).
		Return("refresh-token", nil).Once()
	sessionRepo.On("Create", ctx, mock.AnythingOfType("*domain.Session")).Return(nil)

	uc := newSignInUC(uuidProv, userRepo, sessionRepo, hasher, tokenGen)
	out, err := uc.Execute(ctx, input)

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, user.ID, out.UserID)
	assert.Equal(t, "access-token", out.AccessToken)
	assert.Equal(t, "refresh-token", out.RefreshToken)

	userRepo.AssertExpectations(t)
	hasher.AssertExpectations(t)
	uuidProv.AssertExpectations(t)
	tokenGen.AssertExpectations(t)
	sessionRepo.AssertExpectations(t)
}

func TestSignIn_MissingRequiredFields(t *testing.T) {
	cases := []struct {
		name  string
		input usecase.SignInInput
	}{
		{
			name:  "empty phone",
			input: usecase.SignInInput{Password: "pass"},
		},
		{
			name:  "empty password",
			input: usecase.SignInInput{Phone: "+996700000000"},
		},
		{
			name:  "both empty",
			input: usecase.SignInInput{},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			uuidProv := &mockUUIDProvider{}
			userRepo := &mockUserRepository{}
			sessionRepo := &mockSessionRepository{}
			hasher := &mockPasswordHasher{}
			tokenGen := &mockTokenGenerator{}

			uc := newSignInUC(uuidProv, userRepo, sessionRepo, hasher, tokenGen)
			out, err := uc.Execute(context.Background(), tc.input)

			assert.Nil(t, out)
			assert.EqualError(t, err, "required fields were not filled")
			userRepo.AssertNotCalled(t, "FindByPhone", mock.Anything, mock.Anything)
		})
	}
}

func TestSignIn_UserNotFound(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockUUIDProvider{}
	userRepo := &mockUserRepository{}
	sessionRepo := &mockSessionRepository{}
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}

	input := validSignInInput()
	userRepo.On("FindByPhone", ctx, input.Phone).Return(nil, nil)

	uc := newSignInUC(uuidProv, userRepo, sessionRepo, hasher, tokenGen)
	out, err := uc.Execute(ctx, input)

	assert.Nil(t, out)
	assert.EqualError(t, err, "invalid username or password")
	hasher.AssertNotCalled(t, "Compare", mock.Anything, mock.Anything)
}

func TestSignIn_FindByPhoneError(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockUUIDProvider{}
	userRepo := &mockUserRepository{}
	sessionRepo := &mockSessionRepository{}
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}

	input := validSignInInput()
	userRepo.On("FindByPhone", ctx, input.Phone).Return(nil, errors.New("db error"))

	uc := newSignInUC(uuidProv, userRepo, sessionRepo, hasher, tokenGen)
	out, err := uc.Execute(ctx, input)

	assert.Nil(t, out)
	assert.ErrorContains(t, err, "failed to find a user by phone number")
}

func TestSignIn_WrongPassword(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockUUIDProvider{}
	userRepo := &mockUserRepository{}
	sessionRepo := &mockSessionRepository{}
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}

	input := validSignInInput()
	user := fakeUser()

	userRepo.On("FindByPhone", ctx, input.Phone).Return(user, nil)
	hasher.On("Compare", user.PasswordHash, input.Password).Return(false, nil)

	uc := newSignInUC(uuidProv, userRepo, sessionRepo, hasher, tokenGen)
	out, err := uc.Execute(ctx, input)

	assert.Nil(t, out)
	assert.EqualError(t, err, "incorrect password")
	tokenGen.AssertNotCalled(t, "GenerateToken", mock.Anything, mock.Anything, mock.Anything)
}

func TestSignIn_ComparePasswordError(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockUUIDProvider{}
	userRepo := &mockUserRepository{}
	sessionRepo := &mockSessionRepository{}
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}

	input := validSignInInput()
	user := fakeUser()

	userRepo.On("FindByPhone", ctx, input.Phone).Return(user, nil)
	hasher.On("Compare", user.PasswordHash, input.Password).Return(false, errors.New("hash error"))

	uc := newSignInUC(uuidProv, userRepo, sessionRepo, hasher, tokenGen)
	out, err := uc.Execute(ctx, input)

	assert.Nil(t, out)
	assert.ErrorContains(t, err, "failed to compare passwords")
}

func TestSignIn_AccessTokenGenerationError(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockUUIDProvider{}
	userRepo := &mockUserRepository{}
	sessionRepo := &mockSessionRepository{}
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}

	input := validSignInInput()
	user := fakeUser()

	userRepo.On("FindByPhone", ctx, input.Phone).Return(user, nil)
	hasher.On("Compare", user.PasswordHash, input.Password).Return(true, nil)
	uuidProv.On("Generate").Return("session-uuid-001")
	// First GenerateToken call (access token) fails
	tokenGen.On("GenerateToken", user.ID, "session-uuid-001", mock.Anything).
		Return("", errors.New("signing error")).Once()

	uc := newSignInUC(uuidProv, userRepo, sessionRepo, hasher, tokenGen)
	out, err := uc.Execute(ctx, input)

	assert.Nil(t, out)
	assert.ErrorContains(t, err, "failed to generate access tokens")
	sessionRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestSignIn_RefreshTokenGenerationError(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockUUIDProvider{}
	userRepo := &mockUserRepository{}
	sessionRepo := &mockSessionRepository{}
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}

	input := validSignInInput()
	user := fakeUser()

	userRepo.On("FindByPhone", ctx, input.Phone).Return(user, nil)
	hasher.On("Compare", user.PasswordHash, input.Password).Return(true, nil)
	uuidProv.On("Generate").Return("session-uuid-001")
	// First call succeeds (access token), second fails (refresh token)
	tokenGen.On("GenerateToken", user.ID, "session-uuid-001", mock.Anything).
		Return("access-token", nil).Once()
	tokenGen.On("GenerateToken", user.ID, "session-uuid-001", mock.Anything).
		Return("", errors.New("signing error")).Once()

	uc := newSignInUC(uuidProv, userRepo, sessionRepo, hasher, tokenGen)
	out, err := uc.Execute(ctx, input)

	assert.Nil(t, out)
	assert.ErrorContains(t, err, "failed to generate refresh tokens")
	sessionRepo.AssertNotCalled(t, "Create", mock.Anything, mock.Anything)
}

func TestSignIn_SessionCreateError(t *testing.T) {
	ctx := context.Background()

	uuidProv := &mockUUIDProvider{}
	userRepo := &mockUserRepository{}
	sessionRepo := &mockSessionRepository{}
	hasher := &mockPasswordHasher{}
	tokenGen := &mockTokenGenerator{}

	input := validSignInInput()
	user := fakeUser()

	userRepo.On("FindByPhone", ctx, input.Phone).Return(user, nil)
	hasher.On("Compare", user.PasswordHash, input.Password).Return(true, nil)
	uuidProv.On("Generate").Return("session-uuid-001")
	tokenGen.On("GenerateToken", user.ID, "session-uuid-001", mock.Anything).
		Return("access-token", nil).Once()
	tokenGen.On("GenerateToken", user.ID, "session-uuid-001", mock.Anything).
		Return("refresh-token", nil).Once()
	sessionRepo.On("Create", ctx, mock.AnythingOfType("*domain.Session")).
		Return(errors.New("db write error"))

	uc := newSignInUC(uuidProv, userRepo, sessionRepo, hasher, tokenGen)
	out, err := uc.Execute(ctx, input)

	assert.Nil(t, out)
	assert.ErrorContains(t, err, "failed to establish secure session")
}
