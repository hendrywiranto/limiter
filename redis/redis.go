package redis

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/hendrywiranto/limiter"
	"github.com/redis/go-redis/v9"
)

type redisClient interface {
	Get(ctx context.Context, key string) *redis.StringCmd
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.StatusCmd
	IncrBy(ctx context.Context, key string, value int64) *redis.IntCmd
	MGet(ctx context.Context, keys ...string) *redis.SliceCmd
}

type Adapter struct {
	client redisClient
}

func NewAdapter(client redisClient) *Adapter {
	return &Adapter{client: client}
}

func (a *Adapter) Get(ctx context.Context, key string, value interface{}) error {
	buff, err := a.client.Get(ctx, key).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return limiter.ErrCacheMiss
		}

		return err
	}

	return json.Unmarshal(buff, value)
}

func (a *Adapter) Set(ctx context.Context, key string, value int64, exp time.Duration) error {
	return a.client.Set(ctx, key, value, exp).Err()
}

func (a *Adapter) IncrBy(ctx context.Context, key string, value int64) error {
	_, err := a.client.IncrBy(ctx, key, value).Result()
	return err
}

func (a *Adapter) SumKeys(ctx context.Context, keys []string) (int64, error) {
	res, err := a.client.MGet(ctx, keys...).Result()
	if err != nil {
		return 0, err
	}

	var sum int64
	for _, val := range res {
		if numVal, ok := val.(int64); ok {
			sum += numVal
		}
	}

	return sum, nil
}
