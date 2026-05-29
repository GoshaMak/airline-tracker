package dto

import "api/internal/fleet/domain"

type AircraftModelInfoResponse struct {
	Manufacturer string `json:"manufacturer"`
	Model        string `json:"model"`
	Mass         int    `json:"mass"`
	MaxAltitude  int    `json:"max_altitude"`
	MaxSpeed     int    `json:"max_speed"`
}

func ToResponseAircraftModelInfo(am domain.AircraftModel) (AircraftModelInfoResponse, error) {
	return AircraftModelInfoResponse{
		Manufacturer: am.Manufacturer.String(),
		Model:        am.Model.String(),
		Mass:         am.Mass.Value(),
		MaxAltitude:  am.MaxAltitude.Value(),
		MaxSpeed:     am.MaxSpeed.Value(),
	}, nil
}
