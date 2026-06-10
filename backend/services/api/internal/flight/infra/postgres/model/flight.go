package model

import (
	"time"

	"github.com/google/uuid"
)

type FlightModel struct {
	Id                 uuid.UUID  `db:"id"`
	AircraftId         uuid.UUID  `db:"aircraft_id"`
	ScheduledDeparture time.Time  `db:"scheduled_departure"`
	ScheduledArrival   time.Time  `db:"scheduled_arrival"`
	ActualDeparture    *time.Time `db:"actual_departure"`
	ActualArrival      *time.Time `db:"actual_arrival"`
	Status             string     `db:"status"`
	Plan               *string    `db:"plan"`
	DepartureAirportId uuid.UUID  `db:"departure_airport_id"`
	ArrivalAirportId   uuid.UUID  `db:"arrival_airport_id"`
	DepartureGateId    uuid.UUID  `db:"departure_gate_id"`
	ArrivalGateId      uuid.UUID  `db:"arrival_gate_id"`
}
