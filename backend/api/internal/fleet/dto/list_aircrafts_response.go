package dto

import (
	"api/internal/fleet/domain"
	"api/internal/pagination"

	"github.com/google/uuid"
)

type AircraftResponse struct {
	Id                 uuid.UUID `json:"id"`
	AircraftModelId    uuid.UUID `json:"aircraft_model_id"`
	RegistrationNumber string    `json:"registration_number"`
	SerialNumber       string    `json:"serial_number"`
	Mileage            int       `json:"mileage"`
}

type ListAircraftsResponse struct {
	Items      []AircraftResponse `json:"items"`
	NextCursor *string            `json:"next_cursor"`
	HasMore    bool               `json:"has_more"`
}

func ToResponseListAircrafts(page pagination.Page[domain.Aircraft]) ListAircraftsResponse {
	mapped := pagination.Map(page, func(a domain.Aircraft) AircraftResponse {
		return AircraftResponse{
			Id:                 a.Id,
			AircraftModelId:    a.AircraftModelId,
			RegistrationNumber: a.RegistrationNumber.String(),
			SerialNumber:       a.SerialNumber.String(),
			Mileage:            int(a.Mileage),
		}
	})
	return ListAircraftsResponse{
		Items:      mapped.Items,
		NextCursor: mapped.NextCursor,
		HasMore:    mapped.HasMore,
	}
}
