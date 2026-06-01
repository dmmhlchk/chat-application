package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"identity-service/internal/domain"

	"github.com/google/uuid"
)

var _ domain.UserRepository = (*UserRepository)(nil)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) ExistsByPhoneOrUsername(ctx context.Context, phone, username string) (bool, error) {
	query := `select exists(select 1 from users where username = $1 or phone = $2)`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, username, phone).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("postgres check phone existence failed: %w", err)
	}

	return exists, nil
}

func (r *UserRepository) FindByID(ctx context.Context, userID string) (*domain.User, error) {
	query := `
		select id, username, phone, password_hash 
		from users 
		where id = $1`

	var u domain.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&u)
	if err != nil {
		return nil, fmt.Errorf("postgres find by id failed: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) FindByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
		select id, username, phone, password_hash 
		from users 
		where username = $1`

	var u domain.User
	err := r.db.QueryRowContext(ctx, query, username).Scan(&u)
	if err != nil {
		return nil, fmt.Errorf("postgres find by username failed: %w", err)
	}

	return &u, nil
}

func (r *UserRepository) FindByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
		select id, username, phone, password_hash 
		from users 
		where phone = $1`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, phone).Scan(&user)
	if err != nil {
		return nil, fmt.Errorf("postgres find by username failed: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) NewUUID() string {
	return uuid.New().String()
}

func (r *UserRepository) Create(ctx context.Context, u *domain.User) error {
	query := `
		insert into users (username, password, phone) 
		values ($1, $2, $3) 
		returning id`

	// Scanning back the auto-generated primary key sequence directly onto the domain pointer
	err := r.db.QueryRowContext(ctx, query, u.Username, u.PasswordHash, u.Phone).Scan(&u.ID)
	if err != nil {
		return fmt.Errorf("postgres user insertion failed: %w", err)
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

	_, err := r.db.ExecContext(ctx, query, u.ID, u.Username, u.PasswordHash, u.Phone)
	if err != nil {
		return fmt.Errorf("postgres user update failed: %w", err)
	}

	return nil
}

func (r *UserRepository) Delete(ctx context.Context, userID string) error {
	query := `delete from users where id = $1`

	_, err := r.db.ExecContext(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("postgres user deletion failed: %w", err)
	}

	return nil
}
