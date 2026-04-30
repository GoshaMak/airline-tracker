package model

import (
	"time"

	"github.com/google/uuid"
)

type FlightModel struct {
	ID                 uuid.UUID  `db:"id"`
	AircraftID         uuid.UUID  `db:"aircraft_id"`
	ScheduledDeparture time.Time  `db:"scheduled_departure"`
	ScheduledArrival   time.Time  `db:"scheduled_arrival"`
	ActualDeparture    *time.Time `db:"actual_departure"`
	ActualArrival      *time.Time `db:"actual_arrival"`
	Status             string     `db:"status"`
	Plan               *string    `db:"plan"`
	DepartureAirportID uuid.UUID  `db:"departure_airport_id"`
	ArrivalAirportID   uuid.UUID  `db:"arrival_airport_id"`
	DepartureGateID    uuid.UUID  `db:"departure_gate_id"`
	ArrivalGateID      uuid.UUID  `db:"arrival_gate_id"`
}
