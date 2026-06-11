package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"identity-service/internal/domain"

	"github.com/redis/go-redis/v9"
)

var _ domain.OTPRepository = (*RedisOTPRepository)(nil)

type RedisOTPRepository struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisOTPGenerator(client *redis.Client, ttl time.Duration) *RedisOTPRepository {
	return &RedisOTPRepository{
		client: client,
		ttl:    ttl,
	}
}

func (r *RedisOTPRepository) buildKey(phone string) string {
	return fmt.Sprintf("otp:%s", phone)
}

func (r *RedisOTPRepository) Save(ctx context.Context, phone string, code string) error {
	key := r.buildKey(phone)

	err := r.client.Set(ctx, key, code, r.ttl).Err()
	if err != nil {
		return fmt.Errorf("redis failed to save otp: %w", err)
	}

	return nil
}

func (r *RedisOTPRepository) Verify(ctx context.Context, phone string, code string) (bool, error) {
	key := r.buildKey(phone)

	storedCode, err := r.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return false, nil // Key expired or never existed
		}
		return false, fmt.Errorf("redis failed to fetch otp for verification: %w", err)
	}

	return storedCode == code, nil
}

func (r *RedisOTPRepository) Delete(ctx context.Context, phone string) error {
	key := r.buildKey(phone)

	err := r.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("redis failed to delete otp: %w", err)
	}

	return nil
}
