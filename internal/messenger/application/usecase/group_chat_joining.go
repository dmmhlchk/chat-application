package usecase

import (
	"chat-app/internal/messenger/application/repository"
	"context"
	"fmt"
)

// 1. Determine the input
type GroupJoiningInput struct {
	UserID string
	ChatID string
}

// 2. Determine the dependencies
type GroupJoining struct {
	participantRepo repository.ParticipantRepository
}

func NewGroupJoining(participantRepo repository.ParticipantRepository) *GroupJoining {
	return &GroupJoining{participantRepo: participantRepo}
}

// 3. Business flow of user joining a group chat
func (uc *GroupJoining) Execute(ctx context.Context, input GroupJoiningInput) error {
	// 1. Joining a group chat
	err := uc.participantRepo.Join(ctx, input.ChatID, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to join a group chat: %w", err)
	}

	return nil
}
