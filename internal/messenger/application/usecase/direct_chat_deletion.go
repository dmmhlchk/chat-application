package usecase

import (
	"context"
	"fmt"

	"chat-app/internal/messenger/application/repository"
)

// 1. Determine the input
type DirectDeletionInput struct {
	UserID string
	ChatID string
}

// 2. Determine the dependincies
type DirectDeletion struct {
	chatRepo repository.ChatRepository
}

func NewDirectDeletion(chatRepo repository.ChatRepository) *DirectDeletion {
	return &DirectDeletion{chatRepo: chatRepo}
}

// 3. Business flow of direct chat deletion
func (uc *DirectDeletion) Execute(ctx context.Context, input DirectDeletionInput) error {
	// Check if the chat belongs to the user who wants to delete it
	isMember, err := uc.chatRepo.IsMember(ctx, input.UserID, input.ChatID)
	if err != nil {
		return fmt.Errorf("failed to verify chat: %w", err)
	}
	if !isMember {
		return fmt.Errorf("this chat doesn't belong to this user")
	}

	// Delete it
	if err := uc.chatRepo.Delete(ctx, input.ChatID); err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	return nil
}
