package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"chat-app/internal/identity/application/repository"
	"chat-app/internal/identity/domain"
)

var _ repository.SessionRepository = (*SessionRepository)(nil)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) repository.SessionRepository {
	return &SessionRepository{db: db}
}

// -------------------------------------------------------------------------------------------------
// --		Read methods
// -------------------------------------------------------------------------------------------------

func (r *SessionRepository) FindAllByUserID(ctx context.Context, userID string) ([]domain.Session, error) {
	query := `
		select 
			id, user_id, 
			refresh_token, notification_token, 
		    device_hash, device_name, device_version, device_platform, 
		    created_ip_address, active_ip_address, 
			created_at, active_at, 
			expires_at, is_revoked
		from sessions 
		where user_id = $1
			and is_revoked = false
			and expires > now()
		order by active_at desc`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("postgres: find all sessions by user id failed - %w", err)
	}

	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		var s domain.Session
		err = rows.Scan(
			&s.ID, &s.UserID,
			&s.RefreshTokenHash, &s.NotificationToken,
			&s.Device.Hash, &s.Device.Name, &s.Device.Version, &s.Device.Platform,
			&s.CreatedIPAddress, &s.ActiveIPAddress,
			&s.CreatedAt, &s.ActiveAt,
			&s.ExpiresAt, &s.IsRevoked,
		)
		if err != nil {
			return nil, fmt.Errorf("postgres: scanning session row failed - %w", err)
		}

		sessions = append(sessions, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres: rows iteration error - %w", err)
	}

	return sessions, nil
}

func (r *SessionRepository) FindBySessionID(ctx context.Context, sessionID string) (*domain.Session, error) {
	query := `
		select 
			id, user_id, 
			refresh_token, notification_token, 
		    device_hash, device_name, device_version, device_platform, 
		    created_ip_address, active_ip_address, 
			created_at, active_at, 
			expires_at, is_revoked
		from sessions 
		where id = $1
			and is_revoked = false
			and expires_at > now()`

	var s domain.Session
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&s.ID, &s.UserID,
		&s.RefreshTokenHash, &s.NotificationToken,
		&s.Device.Hash, &s.Device.Name, &s.Device.Version, &s.Device.Platform,
		&s.CreatedIPAddress, &s.ActiveIPAddress,
		&s.CreatedAt, &s.ActiveAt,
		&s.ExpiresAt, &s.IsRevoked,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}

		return nil, fmt.Errorf("postgres: find session by id failed - %w", err)
	}

	return &s, nil
}

// -------------------------------------------------------------------------------------------------
// --		Write methods
// -------------------------------------------------------------------------------------------------

func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	query := `
		insert into sessions 
		(
			user_id, 
			refresh_token, 
			notification_token, 
			device_hash, 
			device_name, 
			device_version, 
			device_platform, 
			created_ip_address, 
			active_ip_address,
			expires_at
		) 
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err := r.db.ExecContext(
		ctx,
		query,
		s.UserID,
		s.RefreshTokenHash,
		s.NotificationToken,
		s.Device.Hash,
		s.Device.Name,
		s.Device.Version,
		s.Device.Platform,
		s.CreatedIPAddress,
		s.ActiveIPAddress,
		s.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("postgres: session insertion failed - %w", err)
	}

	return nil
}

func (r *SessionRepository) Update(ctx context.Context, s *domain.Session) error {
	query := `
		update sessions
		set
			refresh_token = $2, 
			notification_token = $3, 
			device_hash = $4, 
			device_name = $5, 
			device_version = $6, 
			device_platform = $7, 
			active_ip_address = $8,
			active_at = $9,
			expires_at = $10,
			is_revoked = $11
		where id = $1`

	result, err := r.db.ExecContext(
		ctx,
		query,
		s.ID,
		s.RefreshTokenHash,
		s.NotificationToken,
		s.Device.Hash,
		s.Device.Name,
		s.Device.Version,
		s.Device.Platform,
		s.ActiveIPAddress,
		s.ActiveAt,
		s.ExpiresAt,
		s.IsRevoked,
	)

	if err != nil {
		return fmt.Errorf("postgres: session modification failed: %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}

func (r *SessionRepository) TerminateAllByUserID(ctx context.Context, userID string) error {
	query := `
		update sessions
		set is_revoked = true
		where user_id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)

	if err != nil {
		return fmt.Errorf("postgres: terminate all session by user id failed - %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrAlreadyCleanSessions
	}

	return nil
}

func (r *SessionRepository) TerminateBySessionID(ctx context.Context, sessionID string) error {
	query := `
		update sessions
		set is_revoked = true
		where id = $1`

	result, err := r.db.ExecContext(ctx, query, sessionID)

	if err != nil {
		return fmt.Errorf("postgres: terminate session by id failed - %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrSessionNotFound
	}

	return nil
}
