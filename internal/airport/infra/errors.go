package infra

import "errors"

var (
	ErrAirportAlreadyExists = errors.New("airport already exists")
	ErrGateAlreadyExists    = errors.New("gate already exists")
)
