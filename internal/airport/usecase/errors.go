package usecase

import "errors"

var (
	ErrAirportAlreadyExists = errors.New("airport already exists")
	ErrGateAlreadyExists    = errors.New("gate already exists")
	ErrUnexpected           = errors.New("unexpected")
)
