package domain_test

import (
	"identity-service/internal/domain"
	"testing"
	"time"
)

func TestSession_IsValid(t *testing.T) {
	device := domain.Device{
		Hash:     "hash123",
		Name:     "MacBook Pro",
		Version:  "macOS 14.0",
		Platform: "macos",
	}

	tests := []struct {
		name           string
		ttl            time.Duration
		setupSession   func(s *domain.Session)
		expectedResult bool
	}{
		{
			name:           "Valid active session",
			ttl:            1 * time.Hour,
			setupSession:   func(s *domain.Session) {},
			expectedResult: true,
		},
		{
			name:           "Invalid when session is explicitly revoked",
			ttl:            1 * time.Hour,
			setupSession:   func(s *domain.Session) { s.Revoke() },
			expectedResult: false,
		},
		{
			name:           "Invalid when session has expired",
			ttl:            -1 * time.Hour,
			setupSession:   func(s *domain.Session) {},
			expectedResult: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := domain.NewSession(1, "qwerty_refresh_token", "qwerty_notify_token", device, "127.0.0.1", tt.ttl)
			tt.setupSession(session)

			result := session.IsValid()

			if result != tt.expectedResult {
				t.Errorf("expected IsValid() to be %v, got %v", tt.expectedResult, result)
			}

		})
	}
}
