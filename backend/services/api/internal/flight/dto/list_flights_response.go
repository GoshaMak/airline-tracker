package dto

import (
	"api/internal/flight/domain"
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
	Flights []FlightInfo `json:"flights"`
}

func ToResponseListFlights(
	flights []domain.Flight,
) (ListFlightsResponse, error) {
	resp := ListFlightsResponse{
		Flights: make([]FlightInfo, len(flights), cap(flights)),
	}
	for i := range flights {
		resp.Flights[i] = ToFlightInfoDomain(&flights[i])
	}
	return resp, nil
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
