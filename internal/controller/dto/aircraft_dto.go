package dto

import "airline-tracker/internal/domain"

type AircraftDTO struct {
	RegistrationNumber string `json:"registration_number"`
	AircraftModelID    uint   `json:"aircraft_model_id"`
	SerialNumber       string `json:"serial_number"`
	Mileage            uint   `json:"mileage"`
}

func (a *AircraftDTO) AircraftFromDTO() *domain.Aircraft {
	return &domain.Aircraft{
		RegistrationNumber: a.RegistrationNumber,
		AircraftModelID:    a.AircraftModelID,
		SerialNumber:       a.SerialNumber,
		Mileage:            a.Mileage,
	}
}
