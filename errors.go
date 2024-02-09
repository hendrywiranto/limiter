package limiter

import "errors"

var (
	ErrCacheMiss     = errors.New("cache: key not found")
	ErrLimitExceeded = errors.New("limiter: limit exceeded")
)
