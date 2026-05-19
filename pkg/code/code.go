package code

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	mockCode = "123456"
	ttl      = 5 * time.Minute
)

type Service struct {
	rdb *redis.Client
}

func NewService(rdb *redis.Client) *Service {
	return &Service{rdb: rdb}
}

func key(target string) string {
	return fmt.Sprintf("im:vcode:%s", target)
}

// Send stores the verification code for target (email or phone).
// TODO: replace mockCode with real SMS/SMTP sending before production.
func (s *Service) Send(ctx context.Context, target string) error {
	return s.rdb.Set(ctx, key(target), mockCode, ttl).Err()
}

// Verify checks the submitted code and deletes it on success (one-time use).
func (s *Service) Verify(ctx context.Context, target, submitted string) error {
	stored, err := s.rdb.Get(ctx, key(target)).Result()
	if errors.Is(err, redis.Nil) {
		return errors.New("code not found or expired — send code first")
	}
	if err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	if stored != submitted {
		return errors.New("code mismatch")
	}
	s.rdb.Del(ctx, key(target))
	return nil
}
