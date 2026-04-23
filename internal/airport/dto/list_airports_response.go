package dto

import "github.com/google/uuid"

type AirportResponse struct {
	ID uuid.UUID `json:"id"`
	AirportDTO
}

type ListAirportsResponse struct {
	Airports []AirportResponse `json:"airports"`
}
