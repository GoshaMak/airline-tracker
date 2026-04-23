package dto

import (
	"airline-tracker/internal/flight/domain"
	"time"

	"github.com/google/uuid"
)

type flight struct {
	ID                 uuid.UUID  `json:"id"`
	AircraftID         uuid.UUID  `json:"aircraft_id" db:"aircraft_id"`
	ScheduledDeparture time.Time  `json:"scheduled_departure"`
	ScheduledArrival   time.Time  `json:"scheduled_arrival"`
	ActualDeparture    *time.Time `json:"actual_departure"`
	ActualArrival      *time.Time `json:"actual_arrival"`
	Status             string     `json:"status"`
	Plan               string     `json:"plan"`
	DepartureAirportID uuid.UUID  `json:"departure_airport_id"`
	ArrivalAirportID   uuid.UUID  `json:"arrival_airport_id"`
	DepartureGateID    uuid.UUID  `json:"departure_gate_id"`
	ArrivalGateID      uuid.UUID  `json:"arrival_gate_id"`
}

type ListFlightsResponse struct {
	Flights []flight `json:"flights"`
}

func ToResponseListFlights(
	flights []domain.Flight,
) (ListFlightsResponse, error) {
	resp := ListFlightsResponse{
		Flights: make([]flight, len(flights), cap(flights)),
	}
	for i := range flights {
		resp.Flights[i] = domainToResponse(&flights[i])
	}
	return resp, nil
}

func domainToResponse(f *domain.Flight) flight {
	return flight{
		ID:                 f.ID,
		AircraftID:         f.AircraftID,
		ScheduledDeparture: f.ScheduledDeparture,
		ScheduledArrival:   f.ScheduledArrival,
		ActualDeparture:    f.ActualDeparture,
		ActualArrival:      f.ActualArrival,
		Status:             f.Status.String(),
		Plan:               f.Plan.String(),
		DepartureAirportID: f.DepartureAirportID,
		ArrivalAirportID:   f.ArrivalAirportID,
		DepartureGateID:    f.DepartureGateID,
		ArrivalGateID:      f.ArrivalGateID,
	}
}
