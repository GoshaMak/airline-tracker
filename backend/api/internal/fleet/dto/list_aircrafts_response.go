package dto

import (
	"api/internal/fleet/domain"
	"api/internal/pagination"

	"github.com/google/uuid"
)

type aircraft struct {
	Id                 uuid.UUID `json:"id"`
	AircraftModelId    uuid.UUID `json:"aircraft_model_id"`
	RegistrationNumber string    `json:"registration_number"`
	SerialNumber       string    `json:"serial_number"`
	Mileage            int       `json:"mileage"`
}

func ToResponseListAircrafts(page pagination.Page[domain.Aircraft]) pagination.Page[aircraft] {
	return pagination.Map(page, func(a domain.Aircraft) aircraft {
		return aircraft{
			Id:                 a.Id,
			AircraftModelId:    a.AircraftModelId,
			RegistrationNumber: a.RegistrationNumber.String(),
			SerialNumber:       a.SerialNumber.String(),
			Mileage:            int(a.Mileage),
		}
	})
}
