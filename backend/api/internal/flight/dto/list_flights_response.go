package dto

import (
	"api/internal/flight/domain"
	"api/internal/pagination"
	"api/internal/utils"
	"time"

	"github.com/google/uuid"
)

type FlightInfo struct {
	Id         uuid.UUID `json:"id"`
	AircraftId uuid.UUID `json:"aircraft_id"`

	DepartureAirportId uuid.UUID  `json:"departure_airport_id"`
	DepartureGateId    uuid.UUID  `json:"departure_gate_id"`
	ScheduledDeparture time.Time  `json:"scheduled_departure"`
	ActualDeparture    *time.Time `json:"actual_departure"`

	ArrivalAirportId uuid.UUID  `json:"arrival_airport_id"`
	ArrivalGateId    uuid.UUID  `json:"arrival_gate_id"`
	ScheduledArrival time.Time  `json:"scheduled_arrival"`
	ActualArrival    *time.Time `json:"actual_arrival"`

	Status string  `json:"status"`
	Plan   *string `json:"plan"`
}

type ListFlightsResponse struct {
	Items      []FlightInfo `json:"items"`
	NextCursor *string      `json:"next_cursor"`
	HasMore    bool         `json:"has_more"`
}

func ToResponseListFlights(
	page pagination.Page[domain.Flight],
) ListFlightsResponse {
	mapped := pagination.Map(page, func(f domain.Flight) FlightInfo {
		return ToFlightInfoDomain(&f)
	})
	return ListFlightsResponse{
		Items:      mapped.Items,
		NextCursor: mapped.NextCursor,
		HasMore:    mapped.HasMore,
	}
}

func ToFlightInfoDomain(f *domain.Flight) FlightInfo {
	var plan *string
	if f.Plan != nil {
		plan = utils.Ptr(f.Plan.String())
	}
	return FlightInfo{
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
