package dto

import "github.com/google/uuid"

type AirportResponse struct {
	ID uuid.UUID `json:"id"`
	AirportDTO
}

type ListAirportsResponse struct {
	Items      []AirportResponse `json:"items"`
	NextCursor *string           `json:"next_cursor"`
	HasMore    bool              `json:"has_more"`
}
