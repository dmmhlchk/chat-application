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
	userRepo repository.UserRepository
	chatRepo repository.ChatRepository
}

func NewDirectDeletion(
	userRepo repository.UserRepository,
	chatRepo repository.ChatRepository,
) *DirectDeletion {
	return &DirectDeletion{
		userRepo: userRepo,
		chatRepo: chatRepo,
	}
}

// 3. Business flow of direct chat deletion
func (uc *DirectDeletion) Execute(ctx context.Context, input DirectDeletionInput) error {
	// 1. Check if the chat belongs to the user who wants to delete it
	exists, err := uc.chatRepo.IsMember(ctx, input.UserID, input.ChatID)
	if err != nil {
		return fmt.Errorf("failed to verify chat: %w", err)
	}
	if !exists {
		return fmt.Errorf("this chat doesn't belong to this user")
	}

	// 2. Delete it
	err = uc.chatRepo.Delete(ctx, input.ChatID)
	if err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	return nil
}
