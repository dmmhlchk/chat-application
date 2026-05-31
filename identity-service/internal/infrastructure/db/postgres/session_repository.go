package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"identity-service/internal/domain"
)

var _ domain.SessionRepo = (*SessionRepository)(nil)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) FindByID(ctx context.Context, id int) (*domain.Session, error) {
	query := `
		select 
			id, user_id, 
			refresh_token, notification_token, 
		    device_hash, device_name, device_version, device_platform, 
		    created_ip_address, active_ip_address, 
			created_at, active_at, 
			expires_at, is_revoked
		from sessions 
		where id = $1`

	var s domain.Session
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.UserID,
		&s.RefreshTokenHash, &s.NotificationToken,
		&s.Device.Hash, &s.Device.Name, &s.Device.Version, &s.Device.Platform,
		&s.CreatedIPAddress, &s.ActiveIPAddress,
		&s.CreatedAt, &s.ActiveAt,
		&s.ExpiresAt, &s.IsRevoked,
	)
	if err != nil {
		return nil, fmt.Errorf("postgres find by session id failed: %w", err)
	}

	return &s, nil
}

func (r *SessionRepository) FindAllByUserID(ctx context.Context, userId int) ([]domain.Session, error) {
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

	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("postgres find all sessions failed: %w", err)
	}

	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		var s domain.Session
		err := rows.Scan(
			&s.ID, &s.UserID, &s.RefreshTokenHash, &s.NotificationToken,
			&s.Device.Hash, &s.Device.Name, &s.Device.Version, &s.Device.Platform,
			&s.CreatedIPAddress, &s.ActiveIPAddress, &s.CreatedAt, &s.ActiveAt, &s.ExpiresAt, &s.IsRevoked,
		)
		if err != nil {
			return nil, fmt.Errorf("postgres scanning session row failed: %w", err)
		}
		sessions = append(sessions, s)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("postgres rows iteration error: %w", err)
	}

	return sessions, nil

}

func (r *SessionRepository) Create(ctx context.Context, s *domain.Session) error {
	query := `
		insert into sessions (
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
		s.NotificationToken, // Passes empty string default if not set
		s.Device.Hash,       // Assuming a grouped Device domain struct
		s.Device.Name,
		s.Device.Version,
		s.Device.Platform,  // Drivers automatically convert string to custom ENUM type
		s.CreatedIPAddress, // Works flawlessly with a simple string containing the IP
		s.ActiveIPAddress,
		s.ExpiresAt,
	)

	if err != nil {
		return fmt.Errorf("postgres session insertion failed: %w", err)
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

	_, err := r.db.ExecContext(
		ctx,
		query,
		s.ID,
		s.RefreshTokenHash,
		s.NotificationToken, // Passes empty string default if not set
		s.Device.Hash,       // Assuming a grouped Device domain struct
		s.Device.Name,
		s.Device.Version,
		s.Device.Platform,  // Drivers automatically convert string to custom ENUM type
		s.CreatedIPAddress, // Works flawlessly with a simple string containing the IP
		s.ActiveIPAddress,
		s.ExpiresAt,
		s.IsRevoked,
	)

	if err != nil {
		return fmt.Errorf("postgres session insertion failed: %w", err)
	}

	return nil
}
