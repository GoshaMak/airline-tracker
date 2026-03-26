package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Flight struct {
	ID                 uuid.UUID
	AircraftID         *uuid.UUID
	ScheduledDeparture time.Time
	ScheduledArrival   time.Time
	ActualDeparture    time.Time
	ActualArrival      time.Time
	Status             FlightStatus
	Plan               FlightPlan
	DepartureAirportID *uuid.UUID
	ArrivalAirportID   *uuid.UUID
	DepartureGateID    *uuid.UUID
	ArrivalGateID      *uuid.UUID
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
