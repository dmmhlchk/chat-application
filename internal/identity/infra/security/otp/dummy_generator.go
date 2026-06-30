package otp

import "chat-app/internal/identity/application/security"

// Compile-time interface guard
var _ security.OTPGenerator = (*DummyOTPGenerator)(nil)

type DummyOTPGenerator struct{}

func NewDummyOTPGenerator() *DummyOTPGenerator {
	return &DummyOTPGenerator{}
}

func (d *DummyOTPGenerator) Generate(length int) (string, error) {

	return "123123", nil
}
