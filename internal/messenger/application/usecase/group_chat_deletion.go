package usecase

import (
	"context"
	"fmt"

	"chat-app/internal/messenger/application/repository"
	"chat-app/internal/messenger/domain"
)

// 1. Determine the input
type GroupDeletionInput struct {
	UserID string
	ChatID string
}

// 2. Determine the dependencies
type GroupDeletion struct {
	chatRepo repository.ChatRepository
}

func NewGroupDeletion(chatRepo repository.ChatRepository) *GroupDeletion {
	return &GroupDeletion{chatRepo: chatRepo}
}

// 3. Business flow of Group chat deletion
func (uc *GroupDeletion) Execute(ctx context.Context, input GroupDeletionInput) error {

	// Check user permissions
	hasRight, err := uc.chatRepo.CheckPermissions(ctx, input.ChatID, input.UserID, string(domain.UserPermissionDeleteChat))
	if err != nil {
		return fmt.Errorf("failed to check user permissions: %w", err)
	}
	if !hasRight {
		return domain.ErrHasNoRight
	}

	// Delete the group chat
	if err := uc.chatRepo.Delete(ctx, input.ChatID); err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	return nil
}
