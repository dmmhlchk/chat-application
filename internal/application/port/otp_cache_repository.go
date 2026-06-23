package port

import (
	"context"
	"time"
)

type OTPCacheRepository interface {
	Save(ctx context.Context, phone string, code string, ttl time.Duration) error
	Verify(ctx context.Context, phone string, code string) (bool, error)
	Delete(ctx context.Context, phone string) error
}
