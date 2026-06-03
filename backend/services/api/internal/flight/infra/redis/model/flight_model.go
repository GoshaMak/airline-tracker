package model

import (
	"time"

	"github.com/google/uuid"
)

type FlightModel struct {
	Id                 uuid.UUID  `redis:"-"`
	AircraftId         string     `redis:"aircraft_id"`
	ScheduledDeparture time.Time  `redis:"scheduled_departure"`
	ScheduledArrival   time.Time  `redis:"scheduled_arrival"`
	ActualDeparture    *time.Time `redis:"actual_departure,omitempty"`
	ActualArrival      *time.Time `redis:"actual_arrival,omitempty"`
	Status             string     `redis:"status"`
	Plan               *string    `redis:"plan,omitempty"`
	DepartureAirportId string     `redis:"departure_airport_id"`
	ArrivalAirportId   string     `redis:"arrival_airport_id"`
	DepartureGateId    string     `redis:"departure_gate_id"`
	ArrivalGateId      string     `redis:"arrival_gate_id"`
}
