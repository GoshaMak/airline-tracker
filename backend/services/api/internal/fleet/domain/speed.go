package domain

import (
	"errors"
	"strconv"
)

type AircraftMaxSpeed int // km/h

var (
	ErrNegativeAircraftMaxSpeed = errors.New("")
	ErrInvalidAircraftMaxSpeed  = errors.New("")
)

const (
	maxSpeed = 10_000
)

func IsValidAircraftMaxSpeed(s int) error {
	if s < 0 {
		return ErrNegativeAircraftMaxSpeed
	}
	if s > maxSpeed {
		return ErrInvalidAircraftMaxSpeed
	}
	return nil
}

func NewAircraftMaxSpeed(speed int) (AircraftMaxSpeed, error) {
	if err := IsValidAircraftMaxSpeed(speed); err != nil {
		return -1, err
	}
	return AircraftMaxSpeed(speed), nil
}

func (s AircraftMaxSpeed) String() string {
	return strconv.Itoa(int(s))
}

func (s AircraftMaxSpeed) Value() int {
	return int(s)
}
