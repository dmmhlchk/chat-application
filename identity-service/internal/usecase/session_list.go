package usecase

import (
	"context"
	"fmt"
	"identity-service/internal/domain"
	"time"
)

// 1. Determine the input and the output
type SessionListInput struct {
	UserID                  int
	CurrentRefreshTokenHash string
}

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
type SessionList struct {
	sessionRepo domain.SessionRepo
}

func NewSessionList(sessionRepo domain.SessionRepo) *SessionList {
	return &SessionList{
		sessionRepo: sessionRepo,
	}
}

// 3. Business flow of retrieving a list of user sessions
func (uc *SessionList) Execute(ctx context.Context, input SessionListInput) (*SessionListOutput, error) {
	// Retrieving all active sessions by user id
	domainSessions, err := uc.sessionRepo.FindAllByUserID(ctx, input.UserID)
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
