package port

import "github.com/google/uuid"

type UUIDProvider interface {
	Generate() uuid.UUID
}
