package domain

import "github.com/google/uuid"

type FlightRoute struct {
	Id              uuid.UUID
	FlightId        uuid.UUID
	DepartureGateId uuid.UUID
	ArrivalGateId   uuid.UUID
}

func NewFlightRoute(
	flightId,
	departureGateId,
	arrivalGateId uuid.UUID,
) (FlightRoute, error) {
	return FlightRoute{
		Id:              uuid.New(),
		FlightId:        flightId,
		DepartureGateId: departureGateId,
		ArrivalGateId:   arrivalGateId,
	}, nil
}
