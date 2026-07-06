package usecase

import (
	"chat-app/internal/messenger/application/repository"
	"chat-app/internal/messenger/domain"
	"context"
	"fmt"
)

// 1. Determine the input
type GroupModificationInput struct {
	ChatID string
	Title  string
}

// 2. Determine the dependencies
type GroupModification struct {
	chatRepo repository.ChatRepository
}

func NewGroupModification(chatRepo repository.ChatRepository) *GroupModification {
	return &GroupModification{chatRepo: chatRepo}
}

// 3. Business flow of user Adding a group chat
func (uc *GroupModification) Execute(ctx context.Context, input GroupModificationInput) error {

	// 1. Find a group chat by its ID
	group, err := uc.chatRepo.FindByChatID(ctx, input.ChatID)
	if err != nil {
		return fmt.Errorf("failed to find a group by id: %w", err)
	}
	if group == nil {
		return domain.ErrChatNotFound
	}

	// 2. Re-build the group
	group.Title = input.Title

	// 3. Change the group title
	err = uc.chatRepo.Update(ctx, group)
	if err != nil {
		return fmt.Errorf("failed to modify a group chat: %w", err)
	}

	return nil
}
