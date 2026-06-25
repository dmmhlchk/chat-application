package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"chat-application/internal/application/port"
	"chat-application/internal/domain"
)

var _ port.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) port.UserRepository {
	return &UserRepository{db: db}
}

// -------------------------------------------------------------------------------------------------
// --		Read methods
// -------------------------------------------------------------------------------------------------

func (r *UserRepository) FindByUserID(ctx context.Context, userID string) (*domain.User, error) {
	query := `
		select id, username, phone, password_hash 
		from users 
		where id = $1`

	var u domain.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&u.ID, &u.Username, &u.Phone, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("postgres: find user by id failed - %w", err)
	}

	return &u, nil
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
		select id, username, phone, password_hash 
		from users 
		where phone = $1`

	var u domain.User
	err := r.db.QueryRowContext(ctx, query, phone).Scan(&u.ID, &u.Username, &u.Phone, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("postgres: find user by phone failed - %w", err)
	}

	return &u, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		select id, username, phone, password_hash 
		from users 
		where username = $1`

	var u domain.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(&u.ID, &u.Username, &u.Phone, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}

		return nil, fmt.Errorf("postgres: find user by username failed - %w", err)
	}

	return &u, nil
}

func (r *UserRepository) ExistsByPhoneOrUsername(ctx context.Context, phone string, username string) (bool, error) {
	query := `select exists(select 1 from users where username = $1 or phone = $2)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, username, phone).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("postgres: check user existence failed - %w", err)
	}

	return exists, nil
}

// -------------------------------------------------------------------------------------------------
// --		Write methods
// -------------------------------------------------------------------------------------------------

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		insert into users (id, username, phone, password_hash) 
		values ($1, $2, $3, $4)`

	_, err := r.db.ExecContext(ctx, query, u.ID, u.Username, u.Phone, u.PasswordHash)
	if err != nil {
		return fmt.Errorf("postgres: user insertion failed - %w", err)
	}

	return nil
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	query := `
		update users
		set 
			username = $2,
			phone = $3,
			password_hash = $4,
			updated_at = current_timestamp
		where id = $1`

	result, err := r.db.ExecContext(ctx, query, u.ID, u.Username, u.Phone, u.PasswordHash)
	if err != nil {
		return fmt.Errorf("postgres: user modification failed - %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, userID string) error {
	query := `
		delete from users 
		where id = $1`

	result, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("postgres: user deletion failed - %w", err)
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}
