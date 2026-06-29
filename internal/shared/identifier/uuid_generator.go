package identifier

import "github.com/google/uuid"

type UUIDGenerator struct{}

func NewUUIDGenerator() *UUIDGenerator {
	return &UUIDGenerator{}
}

func (p *UUIDGenerator) Generate() string {
	return uuid.NewString()
}
