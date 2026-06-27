package identifier

import (
	"chat-app/internal/identity/application/generator"

	"github.com/google/uuid"
)

var _ generator.IDGenerator = &UUIDGenerator{}

type UUIDGenerator struct{}

func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

func (p *UUIDGenerator) Generate() string {
	return uuid.NewString()
}
