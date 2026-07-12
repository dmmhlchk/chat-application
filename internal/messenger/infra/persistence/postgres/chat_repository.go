package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"chat-app/internal/messenger/application/repository"
	"chat-app/internal/messenger/domain"

	"github.com/lib/pq"
)

var _ repository.ChatRepository = (*ChatRepository)(nil)

type ChatRepository struct {
	db *sql.DB
}

func NewChatRepository(db *sql.DB) repository.ChatRepository {
	return &ChatRepository{db: db}
}

// __Membership methods _________________________________________________________________
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
		return false, err
	}

	return exists, nil
}

// __Permission methods _________________________________________________________________
func (r *ChatRepository) CheckPermissions(ctx context.Context, chatID string, userID string, permission string) (bool, error) {
	// TODO: implement query

	return true, nil
}

func (r *ChatRepository) GetUserPermissions(ctx context.Context, chatID string, userID string) (string, error) {
	// TODO: implement query

	return "", nil
}

// __Read methods _________________________________________________________________
func (r *ChatRepository) FindByChatID(ctx context.Context, chatID string) (*domain.Chat, error) {
	query := `
		select chat_type, created_at, title
		from chats
		where id = $1`

	var chat domain.Chat
	err := r.db.QueryRowContext(ctx, query, chatID).Scan(&chat.Type, &chat.CreatedAt, &chat.Title)
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

func (r *ChatRepository) FindDirectBetween(ctx context.Context, userIDs ...string) (*domain.Chat, error) {
	query := `
		select c.chat_type, c.created_at, c.title
		from chats c
		inner join participans p1 on p1.chat_id = c.id
			and p1.user_id = $1
		inner join participans p2 on p2.chat_id = c.id
			and p2.user_id = $2
		where type = 'direct'`

	var chat domain.Chat
	err := r.db.QueryRowContext(ctx, query, userIDs[0], userIDs[1]).Scan(&chat.Type, &chat.CreatedAt, &chat.Title)
	if err != nil {
		return nil, err
	}

	return &chat, nil
}

// __Write methods _________________________________________________________________
func (r *ChatRepository) Create(ctx context.Context, chat *domain.Chat) error {
	query := `
		insert into chats (id, type, created_at, title)
		values ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query, chat.ID, chat.Type, chat.CreatedAt, chat.Title)
	if err != nil {
		return err
	}

	return nil
}

func (r *ChatRepository) Update(ctx context.Context, chat *domain.Chat) error {
	query := `
		update chats
		set title = $2
		where id = $1`

	_, err := r.db.ExecContext(ctx, query, chat.ID, chat.Title)
	if err != nil {
		return err
	}

	return nil
}

func (r *ChatRepository) Delete(ctx context.Context, chatID string) error {
	query := `
		delete chats 
		where id = $1`

	result, err := r.db.ExecContext(ctx, query, chatID)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrChatNotFound
	}

	return nil
}

func (r *ChatRepository) Join(ctx context.Context, chatID string, userIDs ...string) error {
	if len(userIDs) == 0 {
		return nil
	}

	// Build a query with multiple inserts: insert into participants(...) values (...), (...), (...)
	var sb strings.Builder
	sb.WriteString("insert into participants (chat_id, tag, user_id) values ")

	defaultTag := "member"

	args := make([]interface{}, 0, len(userIDs)+2)
	args = append(args, chatID, defaultTag) // $1, $2

	placeholder := 2
	for idx, userID := range userIDs {
		if idx > 0 {
			sb.WriteString(", ")
		}
		fmt.Fprintf(&sb, "($1, $2, $%d)", placeholder+1)
		args = append(args, userID)
		placeholder += 1
	}

	_, err := r.db.ExecContext(ctx, sb.String(), args)
	return err

}

func (r *ChatRepository) Leave(ctx context.Context, chatID string, userIDs ...string) error {
	if len(userIDs) == 0 {
		return nil
	}

	query := `
		delete from participants 
		where chat_id = $1
			and user_id = any($2)`

	_, err := r.db.ExecContext(ctx, query, chatID, pq.Array(userIDs))
	return err
}

func (r *ChatRepository) SetOwner(ctx context.Context, chatID string, userID string) error {
	query := `
		update participants
		set tag = 'owner'
		where chat_id = $1
			and user_id = $2`

	_, err := r.db.ExecContext(ctx, query, chatID, userID)
	return err
}
