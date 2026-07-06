package usecase

import (
	"context"
	"errors"
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

	// 1. Check if a chat has been already created between these users
	exists, err := uc.chatRepo.ExistsDirectBetween(ctx, userIDs...)
	if err != nil {
		return nil, fmt.Errorf("failed to verify chat: %w", err)
	}
	if exists {
		return nil, errors.New("failed to create a direct chat: another one already created")
	}

	// 2. Create a direct chat
	chatID := uc.idGen.Generate()

	chat := domain.NewChat(chatID, "direct")
	err = uc.chatRepo.Create(ctx, chat)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}

	return &DirectCreationOutput{ChatID: chatID}, nil
}
