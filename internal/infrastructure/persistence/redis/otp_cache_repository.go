package redis

import (
	"context"
	"errors"
	"fmt"
	"time"

	"chat-application/internal/application/port"
	"chat-application/internal/domain"

	"github.com/redis/go-redis/v9"
)

var _ port.OTPCacheRepository = (*OTPCacheRepository)(nil)

type OTPCacheRepository struct {
	client *redis.Client
}

func NewOTPCacheRepository(client *redis.Client) port.OTPCacheRepository {
	return &OTPCacheRepository{client: client}
}

func (r *OTPCacheRepository) Save(ctx context.Context, phone string, code string, ttl time.Duration) error {
	key := r.clearKey(phone)

	err := r.client.Set(ctx, key, code, ttl).Err()
	if err != nil {
		return fmt.Errorf("redis: set otp by phone and code failed - %w", err)
	}

	return nil
}

func (r *OTPCacheRepository) Verify(ctx context.Context, phone string, code string) (bool, error) {
	key := r.clearKey(phone)

	resCode, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, domain.ErrOTPExpired
		}

		return false, fmt.Errorf("redis: verify otp by phone and code failed - %w", err)
	}
	if resCode != code {
		return false, domain.ErrOTPInvalid
	}

	return true, nil
}

func (r *OTPCacheRepository) Delete(ctx context.Context, phone string) error {
	key := r.clearKey(phone)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis: delete otp by phone failed - %w", err)
	}

	return nil
}

// -------------------------------------------------------------------------------------------------
// --		Helpers
// -------------------------------------------------------------------------------------------------

func (r *OTPCacheRepository) clearKey(phone string) string {
	return fmt.Sprintf("otp:%s", phone)
}
