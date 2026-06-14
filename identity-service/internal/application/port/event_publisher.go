package port

import (
	"context"
	"identity-service/internal/domain"
)

type EventPublisher interface {
	PublishUserCreated(ctx context.Context, evt domain.UserCreated) error
	PublishOTPCreated(ctx context.Context, evt domain.OTPCreated) error
}
