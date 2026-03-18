package dto

import "airline-tracker/internal/fleet/domain"

type AircraftModelDTO struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Mass         uint   `json:"mass"`
	MaxAltitude  uint   `json:"max_altitude"`
	MaxSpeed     uint   `json:"max_speed"`
}

func (a *AircraftModelDTO) AircraftModelFromDTO() *domain.AircraftModel {
	return &domain.AircraftModel{
		Manufacturer: a.Manufacturer,
		Model:        a.Model,
		Mass:         a.Mass,
		MaxAltitude:  a.MaxAltitude,
		MaxSpeed:     a.MaxSpeed,
	}
}
