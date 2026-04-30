package dto

import "github.com/google/uuid"

type CreateFlightRequest struct {
	Flight             FlightDTO `json:"flight"`
	AircraftId         uuid.UUID `json:"aircraft_id" example:"MANUALLY"`
	DepartureAirportId uuid.UUID `json:"departure_airport_id" example:"MANUALLY"`
	ArrivalAirportId   uuid.UUID `json:"arrival_airport_id" example:"MANUALLY"`
	DepartureGateId    uuid.UUID `json:"departure_gate_id" example:"MANUALLY"`
	ArrivalGateId      uuid.UUID `json:"arrival_gate_id" example:"MANUALLY"`
}
