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
type DirectChatCreationInput struct {
	FirstUserID  string
	SecondUserID string
}

type DirectChatCreationOutput struct {
	ChatID string
}

// 2. Determine the dependencies
type DirectChatCreation struct {
	idGen           generator.IDGenerator
	userRepo        repository.UserRepository
	participantRepo repository.ParticipantRepository
	chatRepo        repository.ChatRepository
}

func NewDirectChatCreation(
	idGen generator.IDGenerator,
	userRepo repository.UserRepository,
	participantRepo repository.ParticipantRepository,
	chatRepo repository.ChatRepository,
) *DirectChatCreation {
	return &DirectChatCreation{
		idGen:           idGen,
		userRepo:        userRepo,
		participantRepo: participantRepo,
		chatRepo:        chatRepo,
	}
}

// 3. Business flow of chat creation
func (uc *DirectChatCreation) Execute(ctx context.Context, input DirectChatCreationInput) (*DirectChatCreationOutput, error) {

	userIDs := []string{input.FirstUserID, input.SecondUserID}

	// 1. Check if users exist by their id
	exists, err := uc.userRepo.ExistsByUserIDs(ctx, userIDs...)
	if err != nil {
		return nil, fmt.Errorf("failed to verify users: %w", err)
	}
	if !exists {
		return nil, errors.New("failed to create a direct chat: one of the users was not found")
	}

	// 2. Check if a chat has been already created between these users
	exists, err = uc.chatRepo.ExistsDirectChat(ctx, userIDs...)
	if err != nil {
		return nil, fmt.Errorf("failed to verify chat: %w", err)
	}
	if !exists {
		return nil, errors.New("failed to create a direct chat: another one already created")
	}

	// 3. Create a direct chat
	chatID := uc.idGen.Generate()

	chat := domain.NewChat(chatID, "direct")
	err = uc.chatRepo.Create(ctx, chat)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat: %w", err)
	}

	return &DirectChatCreationOutput{ChatID: chatID}, nil
}
