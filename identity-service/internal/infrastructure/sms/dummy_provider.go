package sms

import (
	"context"
	"fmt"
	"identity-service/internal/domain"
)

var _ domain.SMSProvider = (*DummySMSProvider)(nil)

type DummySMSProvider struct{}

func NewDummySMSProvider() *DummySMSProvider {
	return &DummySMSProvider{}
}

// SendOTP prints the OTP directly to stdout using standard formatting
func (p *DummySMSProvider) SendOTP(ctx context.Context, phone string, code string) error {
	fmt.Println("\n--- [MOCK SMS OUTBOUND] ---")
	fmt.Printf("To:      %s\n", phone)
	fmt.Printf("Code:    %s\n", code)
	fmt.Printf("Message: Your verification code is %s. Valid for a minute.\n", code)
	fmt.Println("---------------------------")

	return nil
}
