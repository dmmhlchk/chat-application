package usecase

import (
	"chat-app/internal/messenger/application/repository"
	"context"
	"fmt"
	"time"
)

// 1. Determine the input and the output
type MessageRetrievalInput struct {
	ChatID string
	Page   int
}

type MessageRetrievalOutput struct {
	Messages []Message
}

type Message struct {
	ID      string
	ChatID  string
	UserID  string
	Content string
	SentAt  time.Time
}

// 2. Determine the dependencies
type MessageRetrieval struct {
	chatRepo    repository.ChatRepository
	messageRepo repository.MessageRepositroy
}

func NewMessageRetrieval(
	chatRepo repository.ChatRepository,
	messageRepo repository.MessageRepositroy,
) *MessageRetrieval {
	return &MessageRetrieval{
		chatRepo:    chatRepo,
		messageRepo: messageRepo,
	}
}

// 3. Business flow of getting chat history
func (uc *MessageRetrieval) Execute(ctx context.Context, input MessageRetrievalInput) (*MessageRetrievalOutput, error) {

	// Getting chat history
	domainMessages, err := uc.messageRepo.GetHistory(ctx, input.ChatID, input.Page)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve chat history: %w", err)
	}

	messages := make([]Message, 0, len(domainMessages))
	for _, m := range domainMessages {
		messages = append(messages, Message{
			ID:      m.ID,
			ChatID:  m.ChatID,
			UserID:  m.UserID,
			Content: m.Content,
			SentAt:  m.SentAt,
		})
	}

	return &MessageRetrievalOutput{Messages: messages}, nil
}
