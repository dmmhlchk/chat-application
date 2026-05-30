package usecase

import (
	"context"
	"fmt"
	"identity-service/internal/domain"
	"time"
)

// 1. Determine Input and Output data
// 1.1. SessionListInput defines the data required to retrieve a list of user sessions
type SessionListInput struct {
	UserID                  int
	CurrentRefreshTokenHash string
}

// 1.2. SessionListOutput defines what data we return to the delivery layer upon success
type SessionItem struct {
	ID        int           `json:"id"`
	Device    domain.Device `json:"device"`
	IPAddress string        `json:"ip_address"`
	IsCurrent bool          `json:"is_current"`
	CreatedAt time.Time     `json:"created_at"`
}

type SessionListOutput struct {
	Sessions []SessionItem `json:"sessions"`
}

// 2. Determine the dependencies
// 2.1. SessionList coordinates domain layer to sign in a user
type SessionList struct {
	sessionRepo domain.SessionRepo
}

// 2.2. NewSessionList is a constructor that handles dependency injection
func NewSessionList(sessionRepo domain.SessionRepo) *SessionList {
	return &SessionList{
		sessionRepo: sessionRepo,
	}
}

// 3. Execute retrieves and formats all active sessions for the authenticated user.
func (sl *SessionList) Execute(ctx context.Context, input SessionListInput) (*SessionListOutput, error) {
	// Retrieving all active sessions by user id
	domainSessions, err := sl.sessionRepo.FindAllByUserId(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve active sessions: %w", err)
	}

	items := make([]SessionItem, 0, len(domainSessions))
	for _, s := range domainSessions {
		isCurrent := s.RefreshTokenHash == input.CurrentRefreshTokenHash

		items = append(items, SessionItem{
			ID:        s.ID,
			Device:    s.Device,
			IPAddress: s.ActiveIPAddress,
			IsCurrent: isCurrent,
			CreatedAt: s.CreatedAt,
		})
	}

	return &SessionListOutput{Sessions: items}, nil
}
