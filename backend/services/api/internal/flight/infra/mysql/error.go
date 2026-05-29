package mysql

import "errors"

var (
	ErrFlightNotFound      = errors.New("flight not found")
	ErrFlightRouteNotFound = errors.New("flight route not found")
)
