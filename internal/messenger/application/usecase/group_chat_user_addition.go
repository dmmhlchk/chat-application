package usecase

import (
	"chat-app/internal/messenger/application/repository"
	"chat-app/internal/messenger/domain"
	"context"
	"fmt"
)

// 1. Determine the input
type GroupUserAddingInput struct {
	GranterID string
	ChatID    string
	UserIDs   []string
}

// 2. Determine the dependencies
type GroupUserAdding struct {
	chatRepo repository.ChatRepository
}

func NewGroupUserAdding(chatRepo repository.ChatRepository) *GroupUserAdding {
	return &GroupUserAdding{chatRepo: chatRepo}
}

// 3. Business flow of user Adding a group chat
func (uc *GroupUserAdding) Execute(ctx context.Context, input GroupUserAddingInput) error {
	// Check user permission
	hasRight, err := uc.chatRepo.CheckPermissions(ctx, input.ChatID, input.GranterID, string(domain.UserPermissionAddUser))
	if err != nil {
		return fmt.Errorf("failed to check user permissions: %w", err)
	}
	if !hasRight {
		return domain.ErrHasNoRight
	}

	// Adding new members to the group
	if err := uc.chatRepo.Join(ctx, input.ChatID, input.UserIDs...); err != nil {
		return fmt.Errorf("failed to join a group chat: %w", err)
	}

	return nil
}
