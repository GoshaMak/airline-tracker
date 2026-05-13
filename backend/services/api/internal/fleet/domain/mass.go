package domain

import (
	"errors"
	"strconv"
)

type AircraftMass int // kg

var (
	ErrNegativeAircraftMass = errors.New("negative aircraft mass")
	ErrInvalidAircraftMass  = errors.New("invalid aircraft mass")
)

const (
	maxAircraftMass = 1_000_000
)

func IsValidAircraftMass(m int) error {
	if m < 0 {
		return ErrNegativeAircraftMass
	}
	if m > maxAircraftMass {
		return ErrInvalidAircraftMass
	}
	return nil
}

func NewAircraftMass(mass int) (AircraftMass, error) {
	if err := IsValidAircraftMass(mass); err != nil {
		return -1, err
	}
	return AircraftMass(mass), nil
}

func (m AircraftMass) String() string {
	return strconv.Itoa(int(m))
}
