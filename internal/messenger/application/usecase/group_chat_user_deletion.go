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
	participantRepo repository.ParticipantRepository
}

func NewGroupUserDeletion(participantRepo repository.ParticipantRepository) *GroupDeletion {
	return &GroupDeletion{participantRepo: participantRepo}
}

// 3. Business flow of user Deletion a group chat
func (uc *GroupUserDeletion) Execute(ctx context.Context, input GroupUserDeletionInput) error {
	// 1. Check user permission
	hasRight, err := uc.participantRepo.CheckPermissions(ctx, input.ChatID, input.GranterID, string(domain.UserPermissionDeleteUser))
	if err != nil {
		return fmt.Errorf("failed to check user permissions: %w", err)
	}
	if !hasRight {
		return domain.ErrHasNoRight
	}

	err = uc.participantRepo.Leave(ctx, input.ChatID, input.UserIDs...)
	if err != nil {
		return fmt.Errorf("failed to leave a group chat: %w", err)
	}

	return nil
}
