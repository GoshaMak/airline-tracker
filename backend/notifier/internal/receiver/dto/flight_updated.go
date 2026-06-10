package dto

import (
	"time"

	"github.com/google/uuid"
)

type FlightUpdatedDTO struct {
	FlightId uuid.UUID `json:"flight_id"`
	Users    []string  `json:"users"`

	DepartureAirportTitle string `json:"departure_airport_title"`
	ArrivalAirportTitle   string `json:"arrival_airport_title"`

	ScheduledDeparture *time.Time `json:"scheduled_departure,omitempty"`
	ActualDeparture    *time.Time `json:"actual_departure,omitempty"`

	ScheduledArrival *time.Time `json:"scheduled_arrival,omitempty"`
	ActualArrival    *time.Time `json:"actual_arrival,omitempty"`

	Status *string `json:"status,omitempty"`
	Plan   *string `json:"plan,omitempty"`
}
