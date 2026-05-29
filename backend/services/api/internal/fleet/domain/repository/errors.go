package repository

import "errors"

var (
	ErrAircraftAlreadyExists      = errors.New("aircraft already exists")
	ErrAircraftModelAlreadyExists = errors.New("aircraft model already exists")
	ErrAircraftModelNotFound      = errors.New("aircraft model does not exist")
)
