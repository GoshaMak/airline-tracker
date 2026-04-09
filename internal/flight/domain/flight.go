package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Flight struct {
	ID                 uuid.UUID    `json:"id"`
	AircraftID         *uuid.UUID   `json:"aircraft_id"`
	ScheduledDeparture time.Time    `json:"scheduled_departure"`
	ScheduledArrival   time.Time    `json:"scheduled_arrival"`
	ActualDeparture    time.Time    `json:"actual_departure"`
	ActualArrival      time.Time    `json:"actual_arrival"`
	Status             FlightStatus `json:"status"`
	Plan               FlightPlan   `json:"plan"`
	DepartureAirportID *uuid.UUID   `json:"departure_airport_id"`
	ArrivalAirportID   *uuid.UUID   `json:"arrival_airport_id"`
	DepartureGateID    *uuid.UUID   `json:"departure_gate_id"`
	ArrivalGateID      *uuid.UUID   `json:"arrival_gate_id"`
}

func NewFlight(
	aircraftID *uuid.UUID,
	scheduledDeparture time.Time,
	scheduledArrival time.Time,
	actualDeparture time.Time,
	actualArrival time.Time,
	flightStatus string,
	flightPlan string,
	departureAirportID *uuid.UUID,
	arrivalAirportID *uuid.UUID,
	departureGateID *uuid.UUID,
	arrivalGateID *uuid.UUID,
) (*Flight, error) {
	if scheduledDeparture.Compare(scheduledArrival) != -1 {
		return nil, fmt.Errorf("departure after arrival")
	}
	status, err := NewFlightStatus(flightStatus)
	if err != nil {
		return nil, err
	}
	plan, err := NewFlightPlan(flightPlan)
	if err != nil {
		return nil, err
	}
	return &Flight{
		ID:                 uuid.New(),
		AircraftID:         aircraftID,
		ScheduledDeparture: scheduledDeparture,
		ScheduledArrival:   scheduledArrival,
		ActualDeparture:    actualDeparture,
		ActualArrival:      actualArrival,
		Status:             status,
		Plan:               plan,
		DepartureAirportID: departureAirportID,
		ArrivalAirportID:   arrivalAirportID,
		DepartureGateID:    departureGateID,
		ArrivalGateID:      arrivalGateID,
	}, nil
}
