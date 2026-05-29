package domain

import (
	"time"

	"github.com/google/uuid"
)

type UpdateFlightInfo struct {
	FlightId uuid.UUID

	ScheduledDeparture *time.Time
	ActualDeparture    *time.Time

	ScheduledArrival *time.Time
	ActualArrival    *time.Time

	Status *string
	Plan   *string
}

func NewUpdateFlightInfo(
	flightId uuid.UUID,
	scheduledDeparture,
	actualDeparture,
	scheduledArrival,
	actualArrival *time.Time,
	status,
	plan *string,
) (UpdateFlightInfo, error) {
	return UpdateFlightInfo{
		FlightId:           flightId,
		ScheduledDeparture: scheduledDeparture,
		ActualDeparture:    actualDeparture,
		ScheduledArrival:   scheduledArrival,
		ActualArrival:      actualArrival,
		Status:             status,
		Plan:               plan,
	}, nil
}
