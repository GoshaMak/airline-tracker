package usecase

import "errors"

var (
	ErrCacheSave      = errors.New("error while saving in cache")
	ErrFlightNotFound = errors.New("filght not found")
)
