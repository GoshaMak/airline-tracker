package model

import "github.com/google/uuid"

type FlightRouteModel struct {
	Id              uuid.UUID `db:"id"`
	FlightId        uuid.UUID `db:"flight_id"`
	DepartureGateId uuid.UUID `db:"departure_gate_id"`
	ArrivalGateId   uuid.UUID `db:"arrival_gate_id"`
}
