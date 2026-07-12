package usecase

import (
	"chat-app/internal/messenger/application/repository"
	"chat-app/internal/messenger/domain"
	"context"
	"fmt"
)

// 1. Determine the input
type GroupUserDeletionInput struct {
	GranterID string
	ChatID    string
	UserIDs   []string
}

// 2. Determine the dependencies
type GroupUserDeletion struct {
	chatRepo repository.ChatRepository
}

func NewGroupUserDeletion(chatRepo repository.ChatRepository) *GroupDeletion {
	return &GroupDeletion{chatRepo: chatRepo}
}

// 3. Business flow of user Deletion a group chat
func (uc *GroupUserDeletion) Execute(ctx context.Context, input GroupUserDeletionInput) error {
	// Check user permission
	hasRight, err := uc.chatRepo.CheckPermissions(ctx, input.ChatID, input.GranterID, string(domain.UserPermissionDeleteUser))
	if err != nil {
		return fmt.Errorf("failed to check user permissions: %w", err)
	}
	if !hasRight {
		return domain.ErrHasNoRight
	}

	// Remove users
	if err := uc.chatRepo.Leave(ctx, input.ChatID, input.UserIDs...); err != nil {
		return fmt.Errorf("failed to remove users from the group: %w", err)
	}

	return nil
}
