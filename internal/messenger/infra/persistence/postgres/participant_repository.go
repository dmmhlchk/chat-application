package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"chat-app/internal/messenger/application/repository"
)

var _ repository.ParticipantRepository = (*ParticipantRepository)(nil)

type ParticipantRepository struct {
	db *sql.DB
}

func NewParticipantRepository(db *sql.DB) repository.ParticipantRepository {
	return &ParticipantRepository{db: db}
}

// -------------------------------------------------------------------------------------------------
// --		Write methods
// -------------------------------------------------------------------------------------------------

func (r *ParticipantRepository) Join(ctx context.Context, chatID string, userIDs ...string) error {
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

func (r *ParticipantRepository) SetOwner(ctx context.Context, chatID string, userID string) error {
	query := `
		update participants
		set tag = 'owner'
		where chat_id = $1
			and user_id = $2`

	_, err := r.db.ExecContext(ctx, query, chatID, userID)
	return err
}
