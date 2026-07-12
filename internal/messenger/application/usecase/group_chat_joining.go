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
	chatRepo repository.ChatRepository
}

func NewGroupJoining(chatRepo repository.ChatRepository) *GroupJoining {
	return &GroupJoining{chatRepo: chatRepo}
}

// 3. Business flow of user joining a group chat
func (uc *GroupJoining) Execute(ctx context.Context, input GroupJoiningInput) error {
	// Joining a group chat
	if err := uc.chatRepo.Join(ctx, input.ChatID, input.UserID); err != nil {
		return fmt.Errorf("failed to join a group chat: %w", err)
	}

	return nil
}
