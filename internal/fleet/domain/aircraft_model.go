package domain

import (
	"fmt"

	"github.com/google/uuid"
)

type AircraftModel struct {
	ID           uuid.UUID
	Manufacturer string
	Model        string
	Mass         int
	MaxAltitude  int
	MaxSpeed     int
}

func NewAircraftModel(
	manufacturer string,
	model string,
	mass int,
	maxAltitude int,
	maxSpeed int,
) (*AircraftModel, error) {
	if manufacturer == "" {
		return nil, fmt.Errorf("invalid manufacturer")
	}
	if model == "" {
		return nil, fmt.Errorf("invalid aircraft model")
	}
	if mass <= 0 {
		return nil, fmt.Errorf("negative mass value")
	}
	if maxAltitude <= 0 {
		return nil, fmt.Errorf("negative max altitude")
	}
	if maxSpeed <= 0 {
		return nil, fmt.Errorf("negative max speed")
	}
	return &AircraftModel{
		ID:           uuid.New(),
		Manufacturer: manufacturer,
		Model:        model,
		Mass:         mass,
		MaxAltitude:  maxAltitude,
		MaxSpeed:     maxSpeed,
	}, nil
}
