package usecase_test

import (
	"chat-app/service/identity/domain"
	"context"
	"time"

	"github.com/stretchr/testify/mock"
)

// _________________________________________________________________________
// __ mockUserReader
// _________________________________________________________________________

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

// _________________________________________________________________________
// __ mockEventPublisher
// _________________________________________________________________________

type mockEventPublisher struct{ mock.Mock }

func (m *mockEventPublisher) PublishUserCreated(ctx context.Context, evt domain.UserCreated) error {
	return m.Called(ctx, evt).Error(0)
}

func (m *mockEventPublisher) PublishOTPCreated(ctx context.Context, evt domain.OTPCreated) error {
	return m.Called(ctx, evt).Error(0)
}

// _________________________________________________________________________
// __ mockOTPGenerator
// _________________________________________________________________________

type mockOTPGenerator struct{ mock.Mock }

func (m *mockOTPGenerator) Generate(length int) (string, error) {
	args := m.Called(length)
	return args.String(0), args.Error(1)
}

// _________________________________________________________________________
// __ mockOTPCacheRepository
// _________________________________________________________________________

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

// _________________________________________________________________________
// __ mockUserRepository
// _________________________________________________________________________

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

// _________________________________________________________________________
// __ mockPasswordHasher
// _________________________________________________________________________

type mockPasswordHasher struct{ mock.Mock }

func (m *mockPasswordHasher) Hash(password string) (string, error) {
	args := m.Called(password)
	return args.String(0), args.Error(1)
}

func (m *mockPasswordHasher) Compare(hashedPassword, password string) (bool, error) {
	args := m.Called(hashedPassword, password)
	return args.Bool(0), args.Error(1)
}

// _________________________________________________________________________
// __ mockIDGenerator
// _________________________________________________________________________

type mockIDGenerator struct{ mock.Mock }

func (m *mockIDGenerator) Generate() string {
	return m.Called().String(0)
}

// _________________________________________________________________________
// __ mockSessionRepository
// _________________________________________________________________________

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

// _________________________________________________________________________
// __ mockTokenGenerator
// _________________________________________________________________________

type mockTokenGenerator struct{ mock.Mock }

func (m *mockTokenGenerator) Generate(userID, sessionID string, ttl time.Duration) (string, error) {
	args := m.Called(userID, sessionID, ttl)
	return args.String(0), args.Error(1)
}

func (m *mockTokenGenerator) Validate(token string) (string, string, error) {
	args := m.Called(token)
	return args.String(0), args.String(1), args.Error(2)
}

// _________________________________________________________________________
// __ mockSessionReader
// _________________________________________________________________________

type mockSessionReader struct{ mock.Mock }

func (m *mockSessionReader) FindAllByUserID(ctx context.Context, userID string) ([]domain.Session, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).([]domain.Session), args.Error(1)
}

func (m *mockSessionReader) FindBySessionID(ctx context.Context, sessionID string) (*domain.Session, error) {
	args := m.Called(ctx, sessionID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Session), args.Error(1)
}

// _________________________________________________________________________
// __ mockSessionWriter
// _________________________________________________________________________

type mockSessionWriter struct{ mock.Mock }

func (m *mockSessionWriter) Create(ctx context.Context, session *domain.Session) error {
	return m.Called(ctx, session).Error(0)
}

func (m *mockSessionWriter) Update(ctx context.Context, session *domain.Session) error {
	return m.Called(ctx, session).Error(0)
}

func (m *mockSessionWriter) TerminateAllByUserID(ctx context.Context, userID string) error {
	return m.Called(ctx, userID).Error(0)
}

func (m *mockSessionWriter) TerminateBySessionID(ctx context.Context, sessionID string) error {
	return m.Called(ctx, sessionID).Error(0)
}
