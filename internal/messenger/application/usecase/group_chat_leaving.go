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
	chatRepo repository.ChatRepository
}

func NewGroupLeaving(chatRepo repository.ChatRepository) *GroupLeaving {
	return &GroupLeaving{chatRepo: chatRepo}
}

// 3. Business flow of user leaving a group chat
func (uc *GroupLeaving) Execute(ctx context.Context, input GroupLeavingInput) error {
	// Leaving a group chat
	if err := uc.chatRepo.Leave(ctx, input.ChatID, input.UserID); err != nil {
		return fmt.Errorf("failed to leave a group chat: %w", err)
	}

	return nil
}
