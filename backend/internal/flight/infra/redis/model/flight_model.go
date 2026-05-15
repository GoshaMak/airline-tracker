package model

import (
	"time"

	"github.com/google/uuid"
)

type Flight struct {
	Id                 uuid.UUID  `redis:"id"`
	AircraftId         uuid.UUID  `redis:"aircraft_id"`
	ScheduledDeparture time.Time  `redis:"scheduled_departure"`
	ScheduledArrival   time.Time  `redis:"scheduled_arrival"`
	ActualDeparture    *time.Time `redis:"actual_departure"`
	ActualArrival      *time.Time `redis:"actual_arrival"`
	Status             string     `redis:"status"`
	Plan               *string    `redis:"plan"`
	DepartureAirportId uuid.UUID  `redis:"departure_airport_id"`
	ArrivalAirportId   uuid.UUID  `redis:"arrival_airport_id"`
	DepartureGateId    uuid.UUID  `redis:"departure_gate_id"`
	ArrivalGateId      uuid.UUID  `redis:"arrival_gate_id"`
}
