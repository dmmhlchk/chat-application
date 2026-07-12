package usecase

import (
	"chat-app/internal/messenger/application/repository"
	"context"
	"fmt"
)

// 1. Determine the input and the output
type GroupUserPermissionsInput struct {
	ChatID string
	UserID string
}

type GroupUserPermissionsOutput struct {
	Tag string
}

// 2. Determine the dependencies
type GroupUserPermissions struct {
	chatRepo repository.ChatRepository
}

func NewGroupUserPermissions(chatRepo repository.ChatRepository) *GroupUserPermissions {
	return &GroupUserPermissions{chatRepo: chatRepo}
}

// 3. Business flow of getting user group's permissions
func (uc *GroupUserPermissions) Execute(ctx context.Context, input GroupUserPermissionsInput) (*GroupUserPermissionsOutput, error) {
	// Get user permissions
	tag, err := uc.chatRepo.GetUserPermissions(ctx, input.ChatID, input.UserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user permissions: %w", err)
	}

	return &GroupUserPermissionsOutput{Tag: tag}, nil
}
