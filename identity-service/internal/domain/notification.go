package domain

import "context"

// SMSProvider sends an sms on phone number
type SMSProvider interface {
	SendOTP(ctx context.Context, phone string, code string) error
}
