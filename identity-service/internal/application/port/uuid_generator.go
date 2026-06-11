package port

import "github.com/google/uuid"

type UUIDGeneratod interface {
	NewUUID() uuid.UUID
}
