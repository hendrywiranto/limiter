package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/go-redis/redis"
	"github.com/hendrywiranto/limiter"
)

type redisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	MGet(ctx context.Context, keys ...string) *redis.SliceCmd
}

type adapter struct {
	client redisClient
}

func (a *adapter) Get(ctx context.Context, key string, value interface{}) error {
	buff, err := a.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return limiter.ErrCacheMiss
		}

		return err
	}

	return json.Unmarshal(buff, value)
}

func (a *adapter) Set(ctx context.Context, key string, value int, exp time.Duration) error {
	return a.client.Set(ctx, key, value, exp).Err()
}

func (a *adapter) SumKeys(ctx context.Context, keys []string) (int, error) {
	res, err := a.client.MGet(ctx, keys...).Result()
	if err != nil {
		return 0, err
	}

	var sum int
	for _, val := range res {
		if numVal, ok := val.(int); ok {
			sum += numVal
		}
	}

	return sum, nil
}
