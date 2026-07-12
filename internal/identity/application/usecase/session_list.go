package usecase

import (
	"context"
	"fmt"
	"time"

	"chat-app/internal/identity/application/repository"
	"chat-app/internal/identity/domain"
)

// 1. Determine the input and the output
type SessionListInput struct {
	UserID              string
	CurrentRefreshToken string
}

type SessionListOutput struct {
	Sessions []Session
}

type Session struct {
	ID        string
	Device    domain.Device
	IPAddress string
	IsCurrent bool
	CreatedAt time.Time
}

// 2. Determine the dependencies
type SessionList struct {
	sessionReader repository.SessionReader
}

func NewSessionList(sessionReader repository.SessionReader) *SessionList {
	return &SessionList{
		sessionReader: sessionReader,
	}
}

// 3. Business flow of retrieving a list of user sessions
func (uc *SessionList) Execute(ctx context.Context, input SessionListInput) (*SessionListOutput, error) {
	// Retrieving all active sessions by user id
	domainSessions, err := uc.sessionReader.FindAllByUserID(ctx, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve active sessions: %w", err)
	}

	items := make([]Session, 0, len(domainSessions))
	for _, s := range domainSessions {
		isCurrent := s.RefreshTokenHash == input.CurrentRefreshToken

		items = append(items, Session{
			ID:        s.ID,
			Device:    s.Device,
			IPAddress: s.ActiveIPAddress,
			IsCurrent: isCurrent,
			CreatedAt: s.CreatedAt,
		})
	}

	return &SessionListOutput{Sessions: items}, nil
}
