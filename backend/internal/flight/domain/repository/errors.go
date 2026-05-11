package repository

import "errors"

var (
	ErrCacheEmpty          = errors.New("cache empty")
	ErrFlightNotFound      = errors.New("flight not found")
	ErrFlightRouteNotFound = errors.New("flight route not found")
)
