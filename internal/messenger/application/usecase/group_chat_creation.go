package usecase

import (
	"context"
	"fmt"

	"chat-app/internal/messenger/application/generator"
	"chat-app/internal/messenger/application/repository"
	"chat-app/internal/messenger/domain"
)

// 1. Determine the input and the output
type GroupCreationInput struct {
	Title   string
	OwnerID string
	UserIDs []string // owner included
}

type GroupCreationOutput struct {
	ChatID string
}

// 2. Determine the dependencies
type GroupCreation struct {
	idGen    generator.IDGenerator
	chatRepo repository.ChatRepository
}

func NewGroupCreation(
	idGen generator.IDGenerator,
	chatRepo repository.ChatRepository,
) *GroupCreation {
	return &GroupCreation{
		idGen:    idGen,
		chatRepo: chatRepo,
	}
}

// 3. Business flow of Group chat creation
func (uc *GroupCreation) Execute(ctx context.Context, input GroupCreationInput) (*GroupCreationOutput, error) {
	chatID := uc.idGen.Generate()
	chat := domain.NewChat(chatID, "Group", domain.WithTitle(input.Title))

	// Create a Group chat
	if err := uc.chatRepo.Create(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}

	// Join all members
	if err := uc.chatRepo.Join(ctx, chatID, input.UserIDs...); err != nil {
		_ = uc.chatRepo.Delete(ctx, chatID)
		return nil, fmt.Errorf("failed to join members: %w", err)
	}

	// Set an Owner of the Group chat
	if err := uc.chatRepo.SetOwner(ctx, chatID, input.OwnerID); err != nil {
		_ = uc.chatRepo.Delete(ctx, chatID)
		return nil, fmt.Errorf("failed to set owner: %w", err)
	}

	return &GroupCreationOutput{ChatID: chatID}, nil
}
