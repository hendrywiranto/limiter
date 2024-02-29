package limiter

import (
	"context"
	"fmt"
	"time"
)

const (
	secondFormat = "20060102150405"
	minuteFormat = "200601021504"
	hourFormat   = "2006010215"
)

// Now returns the current time.
// It is a variable so it can be mocked in the tests.
var Now = time.Now

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

// Record records the metric value.
func (l *Limiter) Record(ctx context.Context, metric string, value int64) error {
	if _, ok := l.limits[metric]; !ok {
		return ErrMetricNotFound
	}

	now := Now()
	if err := l.adapter.IncrBy(ctx, fmt.Sprintf("%s:%s", metric, now.Format(secondFormat)), value); err != nil {
		return err
	}
	if err := l.adapter.IncrBy(ctx, fmt.Sprintf("%s:%s", metric, now.Format(minuteFormat)), value); err != nil {
		return err
	}
	if err := l.adapter.IncrBy(ctx, fmt.Sprintf("%s:%s", metric, now.Format(hourFormat)), value); err != nil {
		return err
	}

	return nil
}

// Check checks if the metric has exceeded the limit.
func (l *Limiter) Check(ctx context.Context, metric string, duration Duration) error {
	if _, ok := l.limits[metric]; !ok {
		return ErrMetricNotFound
	}

	keys := l.GenerateKeys(duration)
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

// GenerateKeys generates the keys for the given duration.
func (l *Limiter) GenerateKeys(duration Duration) []string {
	keys := make([]string, 0)
	start := Now().Add(time.Duration(-duration.Seconds()) * time.Second)

	switch duration {
	case DurationSecond:
		keys = append(keys, start.Format(secondFormat))
	case DurationMinute:
		for i := 0; i < 60; i++ {
			key := start.Add(time.Duration(i) * time.Second).Format(secondFormat)
			keys = append(keys, key)
		}
	case DurationHour:
		for i := 0; i < 60-start.Second(); i++ {
			key := start.Add(time.Duration(i) * time.Second).Format(secondFormat)
			keys = append(keys, key)
		}
		for i := 1; i <= 59; i++ {
			key := start.Add(time.Duration(i) * time.Minute).Format(minuteFormat)
			keys = append(keys, key)
		}
		for i := start.Second(); i >= 1; i-- {
			key := start.Add(60 * time.Minute).Add(-time.Duration(i) * time.Second).Format(secondFormat)
			keys = append(keys, key)
		}
	case DurationDay:
		for i := 0; i < 60-start.Second(); i++ {
			key := start.Add(time.Duration(i) * time.Second).Format(secondFormat)
			keys = append(keys, key)
		}
		for i := 1; i < 60-start.Minute(); i++ {
			key := start.Add(time.Duration(i) * time.Minute).Format(minuteFormat)
			keys = append(keys, key)
		}
		for i := 1; i <= 23; i++ {
			key := start.Add(time.Duration(i) * time.Hour).Format(hourFormat)
			keys = append(keys, key)
		}
		for i := start.Minute(); i >= 1; i-- {
			key := start.Add(24 * time.Hour).Add(-time.Duration(i) * time.Minute).Format(minuteFormat)
			keys = append(keys, key)
		}
		for i := start.Second(); i >= 1; i-- {
			key := start.Add(24 * time.Hour).Add(-time.Duration(i) * time.Second).Format(secondFormat)
			keys = append(keys, key)
		}
	}

	return keys
}
