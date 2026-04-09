package dto

import (
	"airline-tracker/internal/flight/domain"
	"time"

	"github.com/google/uuid"
)

type flight struct {
	ID                 uuid.UUID `json:"id"`
	ScheduledDeparture time.Time `json:"scheduled_departure"`
	ScheduledArrival   time.Time `json:"scheduled_arrival"`
	ActualDeparture    time.Time `json:"actual_departure"`
	ActualArrival      time.Time `json:"actual_arrival"`
	Status             string    `json:"status"`
	Plan               string    `json:"plan"`
}

type ListFlightsResponse struct {
	Flights []flight `json:"flights"`
}

func ToResponseListFlights(
	flights []domain.Flight,
) (*ListFlightsResponse, error) {
	resp := &ListFlightsResponse{
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
		ScheduledDeparture: f.ScheduledDeparture,
		ScheduledArrival:   f.ScheduledArrival,
		ActualDeparture:    f.ActualDeparture,
		ActualArrival:      f.ActualArrival,
		Status:             f.Status.String(),
		Plan:               f.Plan.String(),
	}
}
