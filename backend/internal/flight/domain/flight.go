package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type Flight struct {
	Id                 uuid.UUID
	AircraftId         uuid.UUID
	ScheduledDeparture time.Time
	ScheduledArrival   time.Time
	ActualDeparture    *time.Time
	ActualArrival      *time.Time
	Status             FlightStatus
	Plan               *FlightPlan
	DepartureAirportId uuid.UUID
	ArrivalAirportId   uuid.UUID
	DepartureGateId    uuid.UUID
	ArrivalGateId      uuid.UUID
}

func NewFlight(
	aircraftId uuid.UUID,
	scheduledDeparture time.Time,
	scheduledArrival time.Time,
	actualDeparture *time.Time,
	actualArrival *time.Time,
	flightStatus string,
	flightPlan *string,
	departureAirportId uuid.UUID,
	arrivalAirportId uuid.UUID,
	departureGateId uuid.UUID,
	arrivalGateId uuid.UUID,
) (Flight, error) {
	if scheduledDeparture.Compare(scheduledArrival) != -1 {
		return Flight{}, fmt.Errorf("departure after arrival")
	}
	status, err := NewFlightStatus(flightStatus)
	if err != nil {
		return Flight{}, err
	}
	var plan *FlightPlan
	if flightPlan != nil {
		pv, err := NewFlightPlan(*flightPlan)
		if err != nil {
			return Flight{}, err
		}
		plan = &pv
	}
	return Flight{
		Id:                 uuid.New(),
		AircraftId:         aircraftId,
		ScheduledDeparture: scheduledDeparture,
		ScheduledArrival:   scheduledArrival,
		ActualDeparture:    actualDeparture,
		ActualArrival:      actualArrival,
		Status:             status,
		Plan:               plan,
		DepartureAirportId: departureAirportId,
		ArrivalAirportId:   arrivalAirportId,
		DepartureGateId:    departureGateId,
		ArrivalGateId:      arrivalGateId,
	}, nil
}
