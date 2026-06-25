package uuid

import (
	"chat-application/internal/application/port"

	"github.com/google/uuid"
)

var _ port.UUIDProvider = &UUIDProvider{}

type UUIDProvider struct{}

func NewUUIDProvider() *UUIDProvider {
	return &UUIDProvider{}
}

func (p *UUIDProvider) Generate() string {
	return uuid.NewString()
}
