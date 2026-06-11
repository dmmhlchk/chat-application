package port

import "context"

type OTPSender interface {
	Send(ctx context.Context, phone string, code string) error
}
