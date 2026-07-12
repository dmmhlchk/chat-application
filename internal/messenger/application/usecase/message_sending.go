package usecase

import (
	"chat-app/internal/messenger/application/generator"
	"chat-app/internal/messenger/application/repository"
	"chat-app/internal/messenger/domain"
	"context"
	"fmt"
)

// 1. Determine the input and the output
type MessageSendingInput struct {
	UserID  string
	ChatID  string
	Content string
}

type MessageSendingOutput struct {
	MessageID string
}

// 2. Determine the dependencies
type MessageSending struct {
	idGen       generator.IDGenerator
	chatRepo    repository.ChatRepository
	messageRepo repository.MessageRepositroy
}

func NewMessageSending(
	idGen generator.IDGenerator,
	chatRepo repository.ChatRepository,
	messageRepo repository.MessageRepositroy,
) *MessageSending {
	return &MessageSending{
		idGen:       idGen,
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
	}
}

// 3. Business flow of sending a message
func (uc *MessageSending) Execute(ctx context.Context, input MessageSendingInput) (*MessageSendingOutput, error) {
	// Check if the chat belongs to the user who wants to send a message
	isMember, err := uc.chatRepo.IsMember(ctx, input.UserID, input.ChatID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify chat: %w", err)
	}
	if !isMember {
		return nil, fmt.Errorf("this chat doesn't belong to this user")
	}

	// Send a message
	messageID := uc.idGen.Generate()
	message := domain.NewMessage(
		messageID,
		input.ChatID,
		input.UserID,
		domain.MessageTypePlainText,
		input.Content,
	)

	if err := uc.messageRepo.Send(ctx, message); err != nil {
		return nil, fmt.Errorf("failed to send a message: %w", err)
	}

	return &MessageSendingOutput{MessageID: messageID}, nil
}
