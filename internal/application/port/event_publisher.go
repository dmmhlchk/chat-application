package port

import (
	"chat-application/internal/domain"
	"context"
)

type EventPublisher interface {
	PublishUserCreated(ctx context.Context, evt domain.UserCreated) error
	PublishOTPCreated(ctx context.Context, evt domain.OTPCreated) error
}
