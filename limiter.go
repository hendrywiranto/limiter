package limiter

import (
	"context"
	"time"
)

const (
	dayHours = 24
)

type Limiter struct {
	adapter Adapter
	limit   int64
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
		err = l.adapter.Set(ctx, key, 0, dayHours*time.Hour)
		if err != nil {
			return err
		}
	} else {
		return err
	}

	return l.adapter.IncrBy(ctx, key, value)
}

func (l *Limiter) Check(ctx context.Context, metric string, duration Duration) error {
	keys := l.generateKeys(duration)

	sum, err := l.adapter.SumKeys(ctx, keys)
	if err != nil {
		return err
	}

	if sum > l.limit {
		return ErrLimitExceeded
	}

	return nil
}

func (l *Limiter) generateKeys(duration Duration) []string {
	len := duration.Seconds()
	keys := make([]string, len)
	now := time.Now()

	for i := 0; i < int(len); i++ {
		key := now.Add(time.Duration(i) * time.Second).Format("20060102150405")
		keys[i] = key
	}
	return keys
}
