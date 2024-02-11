package limiter

import "errors"

var (
	ErrCacheMiss      = errors.New("cache: key not found")
	ErrLimitExceeded  = errors.New("limiter: limit exceeded")
	ErrLimitNotSet    = errors.New("limiter: limit not set")
	ErrMetricNotFound = errors.New("limiter: metric not found")
)
