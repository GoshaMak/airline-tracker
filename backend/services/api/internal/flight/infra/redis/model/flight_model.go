package model

import (
	"time"

	"github.com/google/uuid"
)

type FlightModel struct {
	Id                 uuid.UUID  `redis:"-"`
	AircraftId         uuid.UUID  `redis:"aircraft_id"`
	ScheduledDeparture time.Time  `redis:"scheduled_departure"`
	ScheduledArrival   time.Time  `redis:"scheduled_arrival"`
	ActualDeparture    *time.Time `redis:"actual_departure,omitempty"`
	ActualArrival      *time.Time `redis:"actual_arrival,omitempty"`
	Status             string     `redis:"status"`
	Plan               *string    `redis:"plan,omitempty"`
	DepartureAirportId uuid.UUID  `redis:"departure_airport_id"`
	ArrivalAirportId   uuid.UUID  `redis:"arrival_airport_id"`
	DepartureGateId    uuid.UUID  `redis:"departure_gate_id"`
	ArrivalGateId      uuid.UUID  `redis:"arrival_gate_id"`
}
