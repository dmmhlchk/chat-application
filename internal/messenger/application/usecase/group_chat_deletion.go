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
	participantRepo repository.ParticipantRepository
	chatRepo        repository.ChatRepository
}

func NewGroupDeletion(
	participantRepo repository.ParticipantRepository,
	chatRepo repository.ChatRepository,
) *GroupDeletion {
	return &GroupDeletion{
		participantRepo: participantRepo,
		chatRepo:        chatRepo,
	}
}

// 3. Business flow of Group chat deletion
func (uc *GroupDeletion) Execute(ctx context.Context, input GroupDeletionInput) error {

	// 1. Check user permissions
	hasRight, err := uc.participantRepo.CheckPermissions(ctx, input.ChatID, input.UserID, string(domain.UserPermissionDeleteChat))
	if err != nil {
		return fmt.Errorf("failed to check user permissions: %w", err)
	}
	if !hasRight {
		return domain.ErrHasNoRight
	}

	// 2. Delete the group chat
	err = uc.chatRepo.Delete(ctx, input.ChatID)
	if err != nil {
		return fmt.Errorf("failed to delete chat: %w", err)
	}

	return nil
}
