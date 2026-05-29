package redis

import "errors"

var (
	ErrCacheEmpty     = errors.New("redis empty")
	ErrFlightNotFound = errors.New("flight not found")
)
