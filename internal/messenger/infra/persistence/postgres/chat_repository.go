package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"chat-app/internal/messenger/application/repository"
	"chat-app/internal/messenger/domain"
)

var _ repository.ChatRepository = (*ChatRepository)(nil)

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) repository.ChatRepository {
	return &ChatRepository{db: db}
}

// -------------------------------------------------------------------------------------------------
// --		Membership methods
// -------------------------------------------------------------------------------------------------

func (r *ChatRepository) IsMember(ctx context.Context, userID string, chatID string) (bool, error) {
	query := `
		select exists
		(
			select 1 
			from participants
			where chat_id = $1
				and user_id = $2
		)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, chatID, userID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("postgres: check membership of chat failed - %w", err)
	}

	return exists, nil
}

// -------------------------------------------------------------------------------------------------
// --		Existence methods
// -------------------------------------------------------------------------------------------------

func (r *ChatRepository) ExistsByChatID(ctx context.Context, chatID string) (bool, error) {
	query := `select exists(select 1 from chats where id = $1)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, chatID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("postgres: check membership of chat failed - %w", err)
	}

	return exists, nil
}

func (r *ChatRepository) ExistsDirectBetween(ctx context.Context, userIDs ...string) (bool, error) {
	if len(userIDs) > 2 {
		return false, fmt.Errorf("was sent more than 2 users")
	}

	query := `
		select exists
		(
			select 1 
			from chats c
			inner join participans p1 on p1.chat_id = c.id
				and p1.user_id = $1
			inner join participans p2 on p2.chat_id = c.id
				and p2.user_id = $2
			where type = 'direct'
		)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, userIDs).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("postgres: check existence of direct chat failed - %w", err)
	}

	return exists, nil
}

// -------------------------------------------------------------------------------------------------
// --		Write methods
// -------------------------------------------------------------------------------------------------

func (r *ChatRepository) Create(ctx context.Context, chat *domain.Chat) error {
	query := `
		insert into chats (id, type, created_at, title)
		values ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query, chat.ID, chat.Type, chat.CreatedAt, chat.Title)
	if err != nil {
		return fmt.Errorf("postgres: chat insertion failed - %w", err)
	}

	return nil
}

func (r *ChatRepository) Delete(ctx context.Context, chatID string) error {
	query := `
		delete chats 
		where id = $1`

	result, err := r.db.ExecContext(ctx, query, chatID)
	if err != nil {
		return fmt.Errorf("postgres: chat insertion failed - %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrChatNotFound
	}

	return nil
}
