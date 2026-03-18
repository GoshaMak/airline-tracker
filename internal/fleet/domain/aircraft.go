package domain

import (
	"errors"

	"github.com/google/uuid"
)

type Aircraft struct {
	id                 uuid.UUID
	RegistrationNumber string
	AircraftModelID    uint
	SerialNumber       string
	Mileage            uint
}

func NewAircraft(
	registrationNumber string,
	aircraftModelID uint,
	serialNumber string,
	mileage uint,
) (*Aircraft, error) {
	if registrationNumber == "" {
		return nil, errors.New("registration number is required")
	}
	if registrationNumber == "" {
		return nil, errors.New("registration number is required")
	}

	return &Aircraft{
		id:                 uuid.New(),
		RegistrationNumber: registrationNumber,
		AircraftModelID:    aircraftModelID,
		SerialNumber:       serialNumber,
		Mileage:            mileage,
	}, nil
}
