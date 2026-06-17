package uuid

import (
	"identity-service/internal/application/port"

	"github.com/google/uuid"
)

var _ port.UUIDProvider = (*UUIDProvider)(nil)

type UUIDProvider struct{}

func NewUUIDProvider() *UUIDProvider {
	return &UUIDProvider{}
}

func (p *UUIDProvider) Generate() string {
	return uuid.NewString()
}
