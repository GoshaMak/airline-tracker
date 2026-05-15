package domain

import (
	"errors"
	"strconv"
)

type AircraftMaxAltitude int // meters

var (
	ErrNegativeAircraftMaxAltitude = errors.New("negative aircraft max altitude")
	ErrInvalidAircraftMaxAltitude  = errors.New("invalid aircraft max altitude")
)

const (
	maxAltitude = 20_000
)

func IsValidAircraftMaxAltitude(a int) error {
	if a < 0 {
		return ErrNegativeAircraftMaxAltitude
	}
	if a > maxAltitude {
		return ErrInvalidAircraftMaxAltitude
	}
	return nil
}

func NewAircraftMaxAltitude(altitude int) (AircraftMaxAltitude, error) {
	if err := IsValidAircraftMaxAltitude(altitude); err != nil {
		return -1, err
	}
	return AircraftMaxAltitude(altitude), nil
}

func (a AircraftMaxAltitude) String() string {
	return strconv.Itoa(int(a))
}
