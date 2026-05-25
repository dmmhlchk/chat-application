package domain

import "context"

type UserRepository interface {
	FindByID(ctx context.Context, id int) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)
	Create(ctx context.Context, user *User)
	Update(ctx context.Context, user *User)
	Delete(ctx context.Context, user *User)
}

type SessionRepository interface {
	Create(ctx context.Context, session *Session)
	Update(ctx context.Context, session *Session)
	Delete(ctx context.Context, session *Session)
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) error
}
