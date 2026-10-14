package repository

import "errors"

var (
	ErrOperationNotPermitted = errors.New("repository operation not permitted")
)
