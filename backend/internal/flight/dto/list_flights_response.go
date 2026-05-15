package dto

import (
	"airline-tracker/internal/flight/domain"
	"airline-tracker/internal/utils"
	"time"

	"github.com/google/uuid"
)

type flight struct {
	Id                 uuid.UUID  `json:"id"`
	AircraftId         uuid.UUID  `json:"aircraft_id" db:"aircraft_id"`
	ScheduledDeparture time.Time  `json:"scheduled_departure"`
	ScheduledArrival   time.Time  `json:"scheduled_arrival"`
	ActualDeparture    *time.Time `json:"actual_departure"`
	ActualArrival      *time.Time `json:"actual_arrival"`
	Status             string     `json:"status"`
	Plan               *string    `json:"plan"`
	DepartureAirportId uuid.UUID  `json:"departure_airport_id"`
	ArrivalAirportId   uuid.UUID  `json:"arrival_airport_id"`
	DepartureGateId    uuid.UUID  `json:"departure_gate_id"`
	ArrivalGateId      uuid.UUID  `json:"arrival_gate_id"`
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
	var plan *string
	if f.Plan != nil {
		plan = utils.Ptr(f.Plan.String())
	}
	return flight{
		Id:                 f.Id,
		AircraftId:         f.AircraftId,
		ScheduledDeparture: f.ScheduledDeparture,
		ScheduledArrival:   f.ScheduledArrival,
		ActualDeparture:    f.ActualDeparture,
		ActualArrival:      f.ActualArrival,
		Status:             f.Status.String(),
		Plan:               plan,
		DepartureAirportId: f.DepartureAirportId,
		ArrivalAirportId:   f.ArrivalAirportId,
		DepartureGateId:    f.DepartureGateId,
		ArrivalGateId:      f.ArrivalGateId,
	}
}
