package domain

import "context"

type UserRepo interface {
	FindByID(ctx context.Context, id int) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByPhone(ctx context.Context, phone string) (*User, error)

	ExistsByPhoneOrUsername(ctx context.Context, phone, username string) (bool, error)

	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, user *User) error
}

type SessionRepo interface {
	FindByUserId(ctx context.Context, userId int) ([]*Session, error)
	FindByToken(ctx context.Context, refreshToken string) (*Session, error)

	Create(ctx context.Context, session *Session) error
	Update(ctx context.Context, session *Session) error
	DeleteByToken(ctx context.Context, refreshToken string) error
}

type PasswordHasher interface {
	Hash(password string) (string, error)
	Compare(hashedPassword, password string) (bool, error)
}

type TokenGenerator interface {
	GeneratePair(int) (string, string, error)
}
