package postgres

import "errors"

type PgErrorCode string

var (
	ErrAircraftAlreadyExists      = errors.New("aircraft already exists")
	ErrAircraftModelAlreadyExists = errors.New("aircraft model already exists")
)
