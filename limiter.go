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
	limits  map[string]Limits
}

// New returns a new Limiter instance.
// adapter is the storage adapter.
// limits is a map of metric name and evaluation duration with its limits.
func New(adapter Adapter, limits map[string]Limits) *Limiter {
	return &Limiter{
		adapter: adapter,
		limits:  limits,
	}
}

func (l *Limiter) Record(ctx context.Context, metric string, value int64) error {
	if _, ok := l.limits[metric]; !ok {
		return ErrMetricNotFound
	}

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
	if _, ok := l.limits[metric]; !ok {
		return ErrMetricNotFound
	}

	keys := l.generateKeys(duration)
	sum, err := l.adapter.SumKeys(ctx, keys)
	if err != nil {
		return err
	}

	if _, ok := l.limits[metric][duration]; !ok {
		return ErrLimitNotSet
	}

	if sum > l.limits[metric][duration] {
		return ErrLimitExceeded
	}

	return nil
}

func (l *Limiter) generateKeys(duration Duration) []string {
	keyLen := duration.Seconds()
	keys := make([]string, keyLen)
	start := time.Now().Add(time.Duration(-keyLen) * time.Second)

	for i := 1; i <= int(keyLen); i++ {
		key := start.Add(time.Duration(i) * time.Second).Format("20060102150405")
		keys[i] = key
	}
	return keys
}
