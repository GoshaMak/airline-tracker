package domain

import "github.com/google/uuid"

type Gate struct {
	Id        uuid.UUID
	AirportId uuid.UUID
	Number    GateNumber
}

func NewGate(aid uuid.UUID, number GateNumber) (Gate, error) {
	return Gate{
		Id:        uuid.New(),
		AirportId: aid,
		Number:    number,
	}, nil
}
