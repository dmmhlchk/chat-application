package postgres

import (
	"chat-app/internal/messenger/application/repository"
	"chat-app/internal/messenger/domain"
	"context"
	"database/sql"
)

var _ repository.MessageRepositroy = (*MessageRepositroy)(nil)

type MessageRepositroy struct {
	db *sql.DB
}

func NewMessageRepositroy(db *sql.DB) repository.MessageRepositroy {
	return &MessageRepositroy{db: db}
}

// __ Read methods _________________________________________________________________
func (r *MessageRepositroy) GetHistory(ctx context.Context, chatID string, page int) ([]domain.Message, error) {
	page_size := 100
	query := `
		select id, chat_id, user_id, type, sent_at, content
		from messages
		where chat_id = $1
		order by sent_at desc
		limit $3 offset ($2 - 1) * $3`

	rows, err := r.db.QueryContext(ctx, query, chatID, max(page, 1), page_size)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var messages []domain.Message
	for rows.Next() {
		var m domain.Message
		err := rows.Scan(&m.ID, &m.ChatID, &m.UserID, &m.Type, &m.SentAt, &m.Content)
		if err != nil {
			return nil, err
		}

		messages = append(messages, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return messages, nil
}

// __ Write methods _________________________________________________________________
func (r *MessageRepositroy) Send(ctx context.Context, message *domain.Message) error {
	query := `
		insert into messages (id, chat_id, user_id, type, sent_at, content)
		values ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query, message.ID, message.ChatID, message.UserID, message.Type, message.SentAt, message.Content)
	if err != nil {
		return err
	}

	return nil
}
