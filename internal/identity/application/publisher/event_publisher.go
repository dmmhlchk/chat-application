package publisher

import (
	"chat-app/internal/identity/domain"
	"context"
)

type EventPublisher interface {
	PublishUserCreated(ctx context.Context, evt domain.UserCreated) error
	PublishOTPCreated(ctx context.Context, evt domain.OTPCreated) error
}
