package dto

import (
	"time"

	"github.com/google/uuid"
)

type FlightDTO struct {
	ID                 uuid.UUID `json:"id"`
	AircraftID         uuid.UUID `json:"aircraft_id"`
	ScheduledDeparture time.Time `json:"scheduled_departure"`
	ScheduledArrival   time.Time `json:"scheduled_arrival"`
	ActualDeparture    time.Time `json:"actual_departure"`
	ActualArrival      time.Time `json:"actual_arrival"`
	Status             string    `json:"status"`
	FlightPlan         string    `json:"flight_plan"`
}

type ListFlightsResponse struct {
	Flights []FlightDTO `json:"flights"`
}
