package dto

import "github.com/google/uuid"

type CreateFlightRequest struct {
	Flight             FlightDTO `json:"flight"`
	AircraftID         uuid.UUID `json:"aircraft_id" example:"MANUALLY"`
	DepartureAirportID uuid.UUID `json:"departure_airport_id" example:"MANUALLY"`
	ArrivalAirportID   uuid.UUID `json:"arrival_airport_id" example:"MANUALLY"`
	DepartureGateID    uuid.UUID `json:"departure_gate_id" example:"MANUALLY"`
	ArrivalGateID      uuid.UUID `json:"arrival_gate_id" example:"MANUALLY"`
}
