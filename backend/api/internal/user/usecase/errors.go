package usecase

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrFlightNotFound      = errors.New("flight not found")
	ErrFlightRouteNotFound = errors.New("flight route not found")
)
