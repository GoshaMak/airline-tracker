package dto

import (
	"time"

	"github.com/google/uuid"
)

type flightUpdateInfo struct {
	FlightId uuid.UUID `json:"id"`

	ScheduledDeparture *time.Time `json:"scheduled_departure"`
	ActualDeparture    *time.Time `json:"actual_departure"`

	ScheduledArrival *time.Time `json:"scheduled_arrival"`
	ActualArrival    *time.Time `json:"actual_arrival"`

	Status *string `json:"status"`
	Plan   *string `json:"plan"`
}

type UpdateFlightRequest struct {
	Flight flightUpdateInfo `json:"flight"`
}
