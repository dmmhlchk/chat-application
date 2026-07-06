package usecase

import (
	"chat-app/internal/messenger/application/repository"
	"context"
	"fmt"
)

// 1. Determine the input
type GroupLeavingInput struct {
	UserID string
	ChatID string
}

// 2. Determine the dependencies
type GroupLeaving struct {
	participantRepo repository.ParticipantRepository
}

func NewGroupLeaving(participantRepo repository.ParticipantRepository) *GroupLeaving {
	return &GroupLeaving{participantRepo: participantRepo}
}

// 3. Business flow of user leaving a group chat
func (uc *GroupLeaving) Execute(ctx context.Context, input GroupLeavingInput) error {
	// 1. Leaving a group chat
	err := uc.participantRepo.Leave(ctx, input.ChatID, input.UserID)
	if err != nil {
		return fmt.Errorf("failed to leave a group chat: %w", err)
	}

	return nil
}
