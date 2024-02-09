package limiter

import (
	"context"
	"time"
)

type Limiter struct {
	adapter Adapter
}

func New(adapter Adapter) *Limiter {
	return &Limiter{
		adapter: adapter,
	}
}

func (l *Limiter) Record(ctx context.Context, metric string, value int64) error {
	now := time.Now().Format("20060102150405")
	key := metric + ":" + now

	if err := l.adapter.Get(ctx, key, nil); err == ErrCacheMiss {
		err = l.adapter.Set(ctx, key, 1, time.Hour)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return l.adapter.IncrBy(ctx, key, value)
}
