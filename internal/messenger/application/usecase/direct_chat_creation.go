package usecase

import (
	"context"
	"fmt"

	"chat-app/internal/messenger/application/generator"
	"chat-app/internal/messenger/application/repository"
	"chat-app/internal/messenger/domain"
)

// 1. Determine the input and the output
type DirectCreationInput struct {
	UserID1 string
	UserID2 string
}

type DirectCreationOutput struct {
	ChatID string
}

// 2. Determine the dependencies
type DirectCreation struct {
	idGen    generator.IDGenerator
	chatRepo repository.ChatRepository
}

func NewDirectCreation(
	idGen generator.IDGenerator,
	chatRepo repository.ChatRepository,
) *DirectCreation {
	return &DirectCreation{
		idGen:    idGen,
		chatRepo: chatRepo,
	}
}

// 3. Business flow of direct chat creation
func (uc *DirectCreation) Execute(ctx context.Context, input DirectCreationInput) (*DirectCreationOutput, error) {

	userIDs := []string{input.UserID1, input.UserID2}

	// Check if a chat has been already created between these users
	chat, err := uc.chatRepo.FindDirectBetween(ctx, userIDs...)
	if err != nil {
		return nil, fmt.Errorf("failed to verify chat: %w", err)
	}
	if chat != nil {
		return &DirectCreationOutput{ChatID: chat.ID}, nil
	}

	// Create a direct chat
	chatID := uc.idGen.Generate()
	chat = domain.NewChat(chatID, "direct")
	if err := uc.chatRepo.Create(ctx, chat); err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}

	// Join all members
	if err := uc.chatRepo.Join(ctx, chatID, userIDs...); err != nil {
		_ = uc.chatRepo.Delete(ctx, chatID)
		return nil, fmt.Errorf("failed to join members: %w", err)
	}

	return &DirectCreationOutput{ChatID: chatID}, nil
}
