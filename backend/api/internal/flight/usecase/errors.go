package usecase

import "errors"

var (
	ErrFlightNotFound        = errors.New("filght not found")
	ErrOperationNotPermitted = errors.New("operation not permitted")
)
