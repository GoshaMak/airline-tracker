package dto

import (
	"api/internal/fleet/domain"

	"github.com/google/uuid"
)

type aircraft struct {
	Id                 uuid.UUID `json:"id"`
	AircraftModelId    uuid.UUID `json:"aircraft_model_id"`
	RegistrationNumber string    `json:"registration_number"`
	SerialNumber       string    `json:"serial_number"`
	Mileage            int       `json:"mileage"`
}

type ListAircraftsResponse struct {
	Aircrafts []aircraft `json:"aircrafts"`
}

func ToResponseListAircrafts(aircrafts []domain.Aircraft) (ListAircraftsResponse, error) {
	resp := ListAircraftsResponse{
		Aircrafts: make([]aircraft, len(aircrafts)),
	}
	for i, a := range aircrafts {
		r := aircraft{
			Id:                 a.Id,
			AircraftModelId:    a.AircraftModelId,
			RegistrationNumber: a.RegistrationNumber.String(),
			SerialNumber:       a.SerialNumber.String(),
			Mileage:            int(a.Mileage),
		}
		resp.Aircrafts[i] = r
	}
	return resp, nil
}
