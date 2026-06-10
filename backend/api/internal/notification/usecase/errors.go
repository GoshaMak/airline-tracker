package usecase

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrFlightNotFound      = errors.New("user flight found")
	ErrFlightRouteNotFound = errors.New("flight route found")
)
