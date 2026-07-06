package repository

import "context"

type ParticipantReader interface {
	CheckPermissions(ctx context.Context, chatID string, userID string, permission string) (bool, error)
}

type ParticipantWriter interface {
	Join(ctx context.Context, chatID string, userIDs ...string) error
	Leave(ctx context.Context, chatID string, userIDs ...string) error
	SetOwner(ctx context.Context, chatID string, userID string) error
}

type ParticipantRepository interface {
	ParticipantReader
	ParticipantWriter
}
