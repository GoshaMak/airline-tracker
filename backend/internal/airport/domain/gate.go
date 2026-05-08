package domain

import "github.com/google/uuid"

type Gate struct {
	ID        uuid.UUID
	AirportID uuid.UUID
	Number    GateNumber
}

func NewGate(aid uuid.UUID, number GateNumber) (Gate, error) {
	return Gate{
		ID:        uuid.New(),
		AirportID: aid,
		Number:    number,
	}, nil
}
