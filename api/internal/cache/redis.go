package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
}

func New(redisURL string) (*Cache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	client := redis.NewClient(opt)
	if err := client.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("ping redis: %w", err)
	}
	return &Cache{client: client}, nil
}

var tierLimits = map[string]int64{
	"free":       100,
	"starter":    10_000,
	"pro":        100_000,
	"enterprise": -1,
}

// CheckLimit increments the daily counter for keyID and returns (allowed, remaining).
func (c *Cache) CheckLimit(ctx context.Context, keyID, tier string) (bool, int64, error) {
	limit, ok := tierLimits[tier]
	if !ok {
		limit = tierLimits["free"]
	}
	if limit == -1 {
		return true, -1, nil
	}

	now := time.Now()
	dayKey := fmt.Sprintf("rl:%s:%s", keyID, now.Format("2006-01-02"))
	tomorrow := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())

	pipe := c.client.TxPipeline()
	incr := pipe.Incr(ctx, dayKey)
	pipe.ExpireAt(ctx, dayKey, tomorrow)
	if _, err := pipe.Exec(ctx); err != nil {
		return false, 0, fmt.Errorf("rate limit pipeline: %w", err)
	}

	count := incr.Val()
	if count > limit {
		return false, 0, nil
	}
	return true, limit - count, nil
}
