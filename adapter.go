package limiter

//go:generate mockgen -source adapter.go -destination mock/adapter.go -package=mock

import (
	"context"
	"time"
)

type Adapter interface {
	Get(ctx context.Context, key string, value interface{}) error
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	IncrBy(ctx context.Context, key string, value int64) error
	SumKeys(ctx context.Context, keys []string) (int64, error)
}
