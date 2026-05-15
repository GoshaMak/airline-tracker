package domain

import (
	"github.com/google/uuid"
)

type AircraftModel struct {
	ID           uuid.UUID
	Manufacturer Manufacturer
	Model        Model
	Mass         AircraftMass
	MaxAltitude  AircraftMaxAltitude
	MaxSpeed     AircraftMaxSpeed
}

func NewAircraftModel(
	manufacturer string,
	model string,
	mass int,
	maxAltitude int,
	maxSpeed int,
) (AircraftModel, error) {
	mnfct, err := NewManufacturer(manufacturer)
	if err != nil {
		return AircraftModel{}, err
	}
	mdl, err := NewModel(model)
	if err != nil {
		return AircraftModel{}, err
	}
	m, err := NewAircraftMass(mass)
	if err != nil {
		return AircraftModel{}, err
	}
	altd, err := NewAircraftMaxAltitude(maxAltitude)
	if err != nil {
		return AircraftModel{}, err
	}
	spd, err := NewAircraftMaxSpeed(maxSpeed)
	if err != nil {
		return AircraftModel{}, err
	}
	return AircraftModel{
		ID:           uuid.New(),
		Manufacturer: mnfct,
		Model:        mdl,
		Mass:         m,
		MaxAltitude:  altd,
		MaxSpeed:     spd,
	}, nil
}
