package dto

import "github.com/google/uuid"

type AircraftDTO struct {
	RegistrationNumber string    `json:"registration_number" example:"999"`
	AircraftModelID    uuid.UUID `json:"aircraft_model_id" example:"fill manually"`
	SerialNumber       string    `json:"serial_number" example:"123"`
	Mileage            int       `json:"mileage" example:"123"`
}
