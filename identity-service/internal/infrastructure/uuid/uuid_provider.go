package uuid

import "github.com/google/uuid"

type UUIDProvider struct{}

func NewUUIDProvider() *UUIDProvider {
	return &UUIDProvider{}
}

func (p *UUIDProvider) Generate() string {
	return uuid.NewString()
}
