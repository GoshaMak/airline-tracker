package usecase

import "errors"

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrFlightNotFound = errors.New("flight not found")
)
