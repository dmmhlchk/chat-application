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
	participantRepo repository.ParticipantRepository
}

func NewGroupUserAdding(participantRepo repository.ParticipantRepository) *GroupUserAdding {
	return &GroupUserAdding{participantRepo: participantRepo}
}

// 3. Business flow of user Adding a group chat
func (uc *GroupUserAdding) Execute(ctx context.Context, input GroupUserAddingInput) error {
	// 1. Check user permission
	hasRight, err := uc.participantRepo.CheckPermissions(ctx, input.ChatID, input.GranterID, string(domain.UserPermissionAddUser))
	if err != nil {
		return fmt.Errorf("failed to check user permissions: %w", err)
	}
	if !hasRight {
		return domain.ErrHasNoRight
	}

	err = uc.participantRepo.Join(ctx, input.ChatID, input.UserIDs...)
	if err != nil {
		return fmt.Errorf("failed to join a group chat: %w", err)
	}

	return nil
}
